// Copyright 2015-2016, Cyrill @ Schumacher.fm and the CoreStore contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package store

import (
	"errors"
	"sync"

	"github.com/corestoreio/csfw/backend"
	"github.com/corestoreio/csfw/config"
	"github.com/corestoreio/csfw/storage/dbr"
	"github.com/corestoreio/csfw/store/scope"
	"github.com/juju/errgo"
)

type (
	// Reader specifies a store Service from which you can only read.
	// A Reader is bound to a scope.Scope.
	Reader interface {
		IsSingleStoreMode() bool
		HasSingleStore() bool
		Website(r ...scope.WebsiteIDer) (*Website, error)
		Websites() (WebsiteSlice, error)
		Group(r ...scope.GroupIDer) (*Group, error)
		Groups() (GroupSlice, error)
		Store(r ...scope.StoreIDer) (*Store, error)
		Stores() (StoreSlice, error)

		// DefaultStoreView because a Reader is bound to a specific scope.Scope,
		// this function will return always the default store view depending on
		// the scope.
		DefaultStoreView() (*Store, error)

		// RequestedStore figures out the default active store for a scope.Option.
		// It takes into account that Reader is bound to a specific scope.Scope.
		// It also prevents running a store from another website or store group,
		// if website or store group was specified explicitly. RequestedStore returns
		// either an error or the store.
		RequestedStore(scope.Option) (activeStore *Store, err error)
	}
)

type (
	// Service represents type which handles the underlying storage and takes care
	// of the default stores. A Service is bound a specific scope.Scope. Depending
	// on the scope it is possible or not to switch stores. A Service contains also
	// a config.Reader which gets passed to the scope of a Store(), Group() or
	// Website() so that you always have the possibility to access a scoped based
	// configuration value.
	// This Service uses three internal maps to cache the pointers
	// of Website, Group and Store.
	Service struct {
		cr config.Getter

		// to which scope is this current Service bound to
		boundToScope scope.Scope

		// storage get set of websites, groups and stores and also type assertion to StorageMutator for
		// ReInit and Persisting
		storage Storager
		mu      sync.RWMutex

		// the next six fields are for internal caching
		// map key is a hash value which is generated by either an int64 or a string.
		// maybe we can get rid of the map by using the existing slices?
		websiteMap map[uint64]*Website
		groupMap   map[uint64]*Group
		storeMap   map[uint64]*Store
		websites   WebsiteSlice
		groups     GroupSlice
		stores     StoreSlice

		// appStore (*cough*) contains the current selected store from init func. Cannot be cleared
		// when booting the app. This store is the main store under which the app runs.
		// In Magento slang it is called currentStore but current Store relates to a Store set
		// by InitByRequest()
		appStore *Store

		// defaultStore some one must be always default.
		defaultStore *Store
	}
)

var _ Reader = (*Service)(nil)

// ErrStoreChangeNotAllowed if a given store within a website would like to
// switch to another store in a different website.
var ErrStoreChangeNotAllowed = errors.New("Store change not allowed")

// ErrHashRetrieverNil internal hashing fails.
var ErrHashRetrieverNil = errors.New("Hash argument is nil")

// NewService creates a new store Service which handles websites, groups and stores.
// A Service can only act on a certain scope (MAGE_RUN_TYPE) and scope ID (MAGE_RUN_CODE).
// Default scope.Scope is always the scope.WebsiteID constant.
// This function is mainly used when booting the app to set the environment configuration
// Also all other calls to any method receiver with nil arguments depends on the internal
// appStore which reflects the default store ID.
func NewService(so scope.Option, storage Storager, opts ...ServiceOption) (*Service, error) {
	scopeID := so.Scope()
	if scopeID == scope.DefaultID {
		scopeID = scope.WebsiteID
	}

	s := &Service{
		cr:           config.DefaultService,
		boundToScope: scopeID,
		storage:      storage,
		mu:           sync.RWMutex{},
		websiteMap:   make(map[uint64]*Website),
		groupMap:     make(map[uint64]*Group),
		storeMap:     make(map[uint64]*Store),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}

	var err error
	s.appStore, err = s.findDefaultStoreByScope(s.boundToScope, so)
	if err != nil {
		if PkgLog.IsDebug() {
			PkgLog.Debug("store.Service.Init", "err", err, "ScopeOption", so)
		}
		return nil, errgo.Mask(err)
	}

	return s, nil
}

// MustNewService same as NewService, but panics on error.
func MustNewService(so scope.Option, storage Storager, opts ...ServiceOption) *Service {
	m, err := NewService(so, storage, opts...)
	if err != nil {
		panic(err)
	}
	return m
}

// findDefaultStoreByScope tries to detect the default store by a given scope option.
// Precedence of detection by passed scope.Option: 1. Store 2. Group 3. Website
func (sm *Service) findDefaultStoreByScope(allowedScope scope.Scope, so scope.Option) (store *Store, err error) {

	switch allowedScope {
	case scope.StoreID:
		store, err = sm.Store(so.Store)
	case scope.GroupID:
		g, errG := sm.Group(so.Group)
		if errG != nil {
			if PkgLog.IsDebug() {
				PkgLog.Debug("store.Service.findDefaultStoreByScope.Group", "err", errG, "ScopeOption", so)
			}
			return nil, errgo.Mask(errG)
		}
		store, err = sm.Store(g) // Group g implements StoreIDer interface to get the default store ID
		if err != nil {
			err = errgo.Mask(ErrGroupDefaultStoreNotFound)
		}
	case scope.WebsiteID:
		if so.Website == nil { // if so.Website == nil then search default website
			store, err = sm.storage.DefaultStoreView() // returns a Store containing less data
			if err == nil {
				store, err = sm.Store(store) // this Store contains more data
			}
		} else {
			w, errW := sm.Website(so.Website)
			if errW != nil {
				if PkgLog.IsDebug() {
					PkgLog.Debug("store.Service.findDefaultStoreByScope.Website", "err", errW, "ScopeOption", so)
				}
				return nil, errgo.Mask(errW)
			}
			g, errG := w.DefaultGroup()
			if errG != nil {
				if PkgLog.IsDebug() {
					PkgLog.Debug("store.Service.findDefaultStoreByScope.Website.DefaultGroup", "err", errG, "ScopeOption", so)
				}
				return nil, errgo.Mask(errG)
			}
			store, err = sm.Store(g) // Group g implements StoreIDer interface to get the default store ID
			if err != nil {
				err = errgo.Mask(ErrGroupDefaultStoreNotFound)
			}
		}
	default:
		err = errgo.Mask(scope.ErrUnsupportedScopeID)
	}
	return
}

// RequestedStore see interface description Reader.RequestedStore
func (sm *Service) RequestedStore(so scope.Option) (activeStore *Store, err error) {

	activeStore, err = sm.findDefaultStoreByScope(so.Scope(), so)
	if err != nil {
		if PkgLog.IsDebug() {
			PkgLog.Debug("store.Service.RequestedStore.FindDefaultStoreByScope", "err", err, "so", so)
		}
		return nil, err
	}

	//	activeStore, err = sm.newActiveStore(activeStore) // this is the active store from a request.
	// todo rethink here if we really need a newActiveStore
	// newActiveStore creates a new Store, Website and Group pointers !!!
	//	if activeStore == nil || err != nil {
	//		// store is not active so ignore
	//		return nil, err
	//	}

	if false == activeStore.Data.IsActive {
		return nil, ErrStoreNotActive
	}

	allowStoreChange := false
	switch sm.boundToScope {
	case scope.StoreID:
		allowStoreChange = true
		break
	case scope.GroupID:
		allowStoreChange = activeStore.Data.GroupID == sm.appStore.Data.GroupID
		break
	case scope.WebsiteID:
		allowStoreChange = activeStore.Data.WebsiteID == sm.appStore.Data.WebsiteID
		break
	}

	if allowStoreChange {
		return activeStore, nil
	}
	return nil, ErrStoreChangeNotAllowed
}

// IsSingleStoreMode check if Single-Store mode is enabled in configuration and from Store count < 3.
// This flag only shows that admin does not want to show certain UI components at backend (like store switchers etc)
// if Magento has only one store view but it does not check the store view collection.
func (sm *Service) IsSingleStoreMode() bool {
	return sm.HasSingleStore() && backend.Backend.GeneralSingleStoreModeEnabled.Get(sm.cr.NewScoped(0, 0, 0)) // default scope
}

// HasSingleStore checks if we only have one store view besides the admin store view.
// Mostly used in models to the set store id and in blocks to not display the store switch.
func (sm *Service) HasSingleStore() bool {
	ss, err := sm.Stores()
	if err != nil {
		return false
	}
	// that means: index 0 is admin store and always present plus one more store view.
	return ss.Len() < 3
}

// Website returns the cached Website pointer from an ID or code including all of its
// groups and all related stores. It panics when the integrity is incorrect.
// If ID and code are available then the non-empty code has precedence.
// If no argument has been supplied then the Website of the internal appStore
// will be returned. If more than one argument has been provided it returns an error.
func (sm *Service) Website(ids ...scope.WebsiteIDer) (*Website, error) {
	lIDs := len(ids)
	emptyIDs := lIDs == 0 || (lIDs == 1 && ids[0] == nil) || lIDs > 1

	if emptyIDs {
		return sm.appStore.Website, nil
	}

	key, err := hash(ids[0], nil, nil)
	if err != nil {
		return nil, err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()
	if w, ok := sm.websiteMap[key]; ok && w != nil {
		return w, nil
	}

	w, err := sm.storage.Website(ids[0])
	sm.websiteMap[key] = w
	return sm.websiteMap[key], errgo.Mask(err)
}

// Websites returns a cached slice containing all pointers to Websites with its associated
// groups and stores. It panics when the integrity is incorrect.
func (sm *Service) Websites() (WebsiteSlice, error) {
	if sm.websites != nil {
		return sm.websites, nil
	}
	var err error
	sm.websites, err = sm.storage.Websites()
	return sm.websites, err
}

// Group returns a cached Group which contains all related stores and its website.
// Only the argument ID is supported.
// If no argument has been supplied then the Group of the internal appStore
// will be returned. If more than one argument has been provided it returns an error.
func (sm *Service) Group(ids ...scope.GroupIDer) (*Group, error) {
	lIDs := len(ids)
	emptyIDs := lIDs == 0 || (lIDs == 1 && ids[0] == nil) || lIDs > 1

	if emptyIDs {
		return sm.appStore.Group, nil
	}

	key, err := hash(nil, ids[0], nil)
	if err != nil {
		return nil, err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()
	if g, ok := sm.groupMap[key]; ok && g != nil {
		return g, nil
	}

	g, err := sm.storage.Group(ids[0])
	sm.groupMap[key] = g
	return sm.groupMap[key], errgo.Mask(err)
}

// Groups returns a cached slice containing all pointers to Groups with its associated
// stores and websites. It panics when the integrity is incorrect.
func (sm *Service) Groups() (GroupSlice, error) {
	if sm.groups != nil {
		return sm.groups, nil
	}
	var err error
	sm.groups, err = sm.storage.Groups()
	return sm.groups, err
}

// Store returns the cached Store view containing its group and its website.
// If ID and code are available then the non-empty code has precedence.
// If no argument has been supplied then the appStore
// will be returned. If more than one argument has been provided it returns an error.
func (sm *Service) Store(ids ...scope.StoreIDer) (*Store, error) {
	lIDs := len(ids)
	emptyIDs := lIDs == 0 || (lIDs == 1 && ids[0] == nil) || lIDs > 1

	if emptyIDs {
		return sm.appStore, nil
	}

	key, err := hash(nil, nil, ids[0])
	if err != nil {
		return nil, err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()
	if s, ok := sm.storeMap[key]; ok && s != nil {
		return s, nil
	}

	s, err := sm.storage.Store(ids[0])
	sm.storeMap[key] = s
	return sm.storeMap[key], errgo.Mask(err)
}

// Stores returns a cached Store slice. Can return an error when the website or
// the group cannot be found.
func (sm *Service) Stores() (StoreSlice, error) {
	if sm.stores != nil {
		return sm.stores, nil
	}
	var err error
	sm.stores, err = sm.storage.Stores()
	return sm.stores, err
}

// DefaultStoreView returns the default store view.
func (sm *Service) DefaultStoreView() (*Store, error) {
	if sm.defaultStore != nil {
		return sm.defaultStore, nil
	}
	var err error
	sm.defaultStore, err = sm.storage.DefaultStoreView()
	return sm.defaultStore, err
}

// newActiveStore returns a new non-cached Store with all its Websites and Groups but only if the Store
// is marked as active. Argument can be an ID or a Code. Returns err if Store not found or inactive.
//func (sm *Service) newActiveStore(r scope.StoreIDer) (*Store, error) {
//	s, err := sm.storage.Store(r)
//	if err != nil {
//		return nil, err
//	}
//	if s.Data.IsActive {
//		return s, nil
//	}
//	return nil, ErrStoreNotActive
//}

// ReInit reloads the website, store group and store view data from the database.
// After reloading internal cache will be cleared if there are no errors.
func (sm *Service) ReInit(dbrSess dbr.SessionRunner, cbs ...dbr.SelectCb) error {
	err := sm.storage.ReInit(dbrSess, cbs...)
	if err == nil {
		sm.ClearCache()
	}
	return err
}

// ClearCache resets the internal caches which stores the pointers to a Website, Group or Store and
// all related slices. Please use with caution. ReInit() also uses this method.
// Providing argument true clears also the internal appStore cache.
func (sm *Service) ClearCache(clearAll ...bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if len(sm.websiteMap) > 0 {
		for k := range sm.websiteMap {
			delete(sm.websiteMap, k)
		}
	}
	if len(sm.groupMap) > 0 {
		for k := range sm.groupMap {
			delete(sm.groupMap, k)
		}
	}
	if len(sm.storeMap) > 0 {
		for k := range sm.storeMap {
			delete(sm.storeMap, k)
		}
	}
	sm.websites = nil
	sm.groups = nil
	sm.stores = nil
	sm.defaultStore = nil
	// do not clear currentStore as this one depends on the init funcs
	if 1 == len(clearAll) && clearAll[0] {
		sm.appStore = nil
	}
}

// IsCacheEmpty returns true if the internal cache is empty.
func (sm *Service) IsCacheEmpty() bool {
	return len(sm.websiteMap) == 0 && len(sm.groupMap) == 0 && len(sm.storeMap) == 0 &&
		sm.websites == nil && sm.groups == nil && sm.stores == nil && sm.defaultStore == nil
}

// hash generates the key for the map from either an id int64 or a code string.
// If both interfaces are nil it returns 0 which is default for website, group or store.
// fnv64a used to calculate the uint64 value of a string, especially website code and store code.
func hash(wID scope.WebsiteIDer, gID scope.GroupIDer, sID scope.StoreIDer) (uint64, error) {
	uz := uint64(0)
	if nil == wID && nil == gID && nil == sID {
		return uz, ErrHashRetrieverNil
	}

	if wC, ok := wID.(scope.WebsiteCoder); ok {
		return hashCode(wC.WebsiteCode()), nil
	}
	if nil != wID {
		return uint64(wID.WebsiteID()), nil
	}

	if nil != gID {
		return uint64(gID.GroupID()), nil
	}

	if sC, ok := sID.(scope.StoreCoder); ok {
		return hashCode(sC.StoreCode()), nil
	}
	if nil != sID {
		return uint64(sID.StoreID()), nil
	}
	return uz, ErrHashRetrieverNil // unreachable ....
}

// fnv hash
func hashCode(code string) uint64 {
	data := []byte(code)
	var hash uint64 = 14695981039346656037
	for _, c := range data {
		hash ^= uint64(c)
		hash *= 1099511628211
	}
	return hash
}
