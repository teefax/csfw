package ctxrouter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/corestoreio/csfw/net/mw"
	"github.com/stretchr/testify/assert"
)

func noopMW() mw.Middleware {
	return func(hf http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hf.ServeHTTP(w, r)
		})
	}
}

func TestGroup(t *testing.T) {
	g := New().Group("/group")

	g.Use(noopMW())

	h := func(http.ResponseWriter, *http.Request) {}
	//g.CONNECT("/", h)
	g.DELETE("/", h)
	g.GET("/", h)
	g.HEAD("/", h)
	g.OPTIONS("/", h)
	g.PATCH("/", h)
	g.POST("/", h)
	g.PUT("/", h)
	g.WEBSOCKET("/ws", h)

	g2 := g.Group("/files")
	mfs := &mockFileSystem{}
	recv := catchPanic(func() {
		g2.ServeFiles("/noFilepath", mfs)
	})
	if recv == nil {
		t.Fatal("registering path not ending with '*filepath' did not panic")
	}
	g2.ServeFiles("/*filepath", mfs)
}

func groupHeader() mw.Middleware {
	return func(hf http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-CoreStore-ID", "Goph3r")
			hf.ServeHTTP(w, r)
		})
	}
}

func TestGroupMiddlewareNoParams(t *testing.T) {
	r := New()
	g := r.Group("/group", groupHeader())
	h := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("Group Error\n"))
	}
	g.GET("/error", h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/group/error", nil)
	r.ServeHTTP(w, req)

	assert.Exactly(t, 500, w.Code)
	assert.Exactly(t, "Group Error\n", w.Body.String())
	assert.Exactly(t, "Goph3r", w.Header().Get("X-CoreStore-ID"), "Header key X-CoreStore-ID not found, which has been applied by a middleware")
}

func mwGroup1() mw.Middleware {
	return func(hf http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ps := FromContextParams(r.Context())
			w.Header().Set("X-CoreStore-ID", "group1")
			w.Header().Set("X-CoreStore-MSG", ps.ByName("msg"))
			hf.ServeHTTP(w, r)
		})
	}
}

func mwGroup2() mw.Middleware {
	return func(hf http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ps := FromContextParams(r.Context())
			w.Header().Set("X-CoreStore-ID", "group2")
			w.Header().Set("X-CoreStore-MSG", ps.ByName("msg"))
			hf.ServeHTTP(w, r)
		})
	}
}

func TestGroupMiddlewareMultipleRoutes(t *testing.T) {

	r := New()

	gh1 := func(w http.ResponseWriter, r *http.Request) {
		ps := FromContextParams(r.Context())
		assert.Exactly(t, "grouperror1", ps.ByName("msg"))
		assert.Exactly(t, "grouperror1", w.Header().Get("X-CoreStore-MSG"), "X-CoreStore-MSG Header not set")
		assert.Exactly(t, "group1", w.Header().Get("X-CoreStore-ID"), "X-CoreStore-ID Header not set")
	}

	r.Use(mwGroup1())
	r.GET("/group/:msg", gh1)

	g1 := r.Group("/group1", mwGroup1())
	g1.GET("/error/:msg", gh1)

	g2 := r.Group("/group2", mwGroup2())
	g2.GET("/error/:msg", func(w http.ResponseWriter, r *http.Request) {
		ps := FromContextParams(r.Context())
		assert.Exactly(t, "grouperror2", ps.ByName("msg"))
		assert.Exactly(t, "grouperror2", w.Header().Get("X-CoreStore-MSG"))
		assert.Exactly(t, "group2", w.Header().Get("X-CoreStore-ID"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/group1/error/grouperror1", nil)
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/group2/error/grouperror2", nil)
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/group/grouperror1", nil)
	r.ServeHTTP(w, req)

}
