package route

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"net/url"
)

func TestRegisterHandlerWithStaticRoute(t *testing.T) {
	// given
	var called bool
	f := func(http.ResponseWriter, *http.Request) {
		called = true
	}
	path := []string{"api", "v1", "team"}

	r := NewDynamicRouter()

	// when
	r.registerHandler(path, f)

	// then
	if len(r.root) != 1 {
		t.Fatal("router must only have one path root")
	}

	api, apiPresent := r.root["api"]
	if !apiPresent {
		t.Fatal("the first node should be on api")
	} else if len(api.children) != 1 {
		t.Fatal("the root node must have one children")
	}

	v1, v1Present := api.children["v1"]
	if !v1Present {
		t.Fatal("the second node should be on v1")
	} else if len(v1.children) != 1 {
		t.Fatal("the second node must have one children")
	}
	team, teamPresent := v1.children["team"]
	if !teamPresent {
		t.Fatal("the last node should be on team")
	} else if len(team.children) != 0 {
		t.Fatal("the last node must have no children")
	}

	team.handler(httptest.NewRecorder(), &http.Request{})
	if !called {
		t.Fatal("the handler has not been correctly registered")
	}
}

func TestRegisterHandlerWithEmptyPartOfPath(t *testing.T) {
	// given
	var called bool
	f := func(http.ResponseWriter, *http.Request) {
		called = true
	}
	path := []string{"api", "v1", "", "team"}

	r := NewDynamicRouter()

	// when
	r.registerHandler(path, f)

	// then
	if len(r.root) != 1 {
		t.Fatal("router must only have one path root")
	}

	api, apiPresent := r.root["api"]
	if !apiPresent {
		t.Fatal("the first node should be on api")
	} else if len(api.children) != 1 {
		t.Fatal("the root node must have one children")
	}

	v1, v1Present := api.children["v1"]
	if !v1Present {
		t.Fatal("the second node should be on v1")
	} else if len(v1.children) != 1 {
		t.Fatal("the second node must have one children")
	}
	team, teamPresent := v1.children["team"]
	if !teamPresent {
		t.Fatal("the last node should be on team")
	} else if len(team.children) != 0 {
		t.Fatal("the last node must have no children")
	}

	team.handler(httptest.NewRecorder(), &http.Request{})
	if !called {
		t.Fatal("the handler has not been correctly registered")
	}
}

func TestRegisterHandlerWithEmptyPath(t *testing.T) {
	// given
	defer func() {
		if r := recover(); r != nil {
			t.Log("successfuly caught the router panic")
		}
	}()
	f := func(http.ResponseWriter, *http.Request) {}

	// at the same tree level, dymanic path value has to be considered as the same resource
	path := []string{}
	r := NewDynamicRouter()

	// when
	r.registerHandler(path, f)

	// the router must panic. If not => fatal
	t.Fatal("expect the router to panic")
}

func TestRegisterHandlerWithNilHandler(t *testing.T) {
	// given
	defer func() {
		if r := recover(); r != nil {
			t.Log("successfuly caught the router panic")
		}
	}()
	// at the same tree level, dymanic path value has to be considered as the same resource
	path := []string{"api", "v1", "item1"}
	r := NewDynamicRouter()

	// when
	r.registerHandler(path, nil)

	// the router must panic. If not => fatal
	t.Fatal("expect the router to panic")
}

func TestRegisterHandlerWithDynamicPartOfPath(t *testing.T) {
	// given

	var called1 bool
	var called2 bool

	f1 := func(http.ResponseWriter, *http.Request) {
		called1 = true
	}

	f2 := func(http.ResponseWriter, *http.Request) {
		called2 = true
	}
	// at the same tree level, dymanic path value has to be considered as the same resource
	path1 := []string{"api", ":someId1", "item1"}
	path2 := []string{"api", ":someId1", "item2"}
	r := NewDynamicRouter()

	// when
	r.registerHandler(path1, f1)
	r.registerHandler(path2, f2)

	// then
	if len(r.root) != 1 {
		t.Fatal("router must only have one path root")
	}

	api, apiPresent := r.root["api"]
	if !apiPresent {
		t.Fatal("the first node should be on api")
	} else if len(api.children) != 1 {
		t.Fatal("the root node must have one children")
	}

	dynamic, dynamicPresent := api.children[":someId1"]
	if !dynamicPresent {
		t.Fatal("the second node should be on v1")
	} else if len(dynamic.children) != 2 {
		t.Fatal("the second node must have two children")
	}

	item1, item1Present := dynamic.children["item1"]
	if !item1Present {
		t.Fatal("the last node should be on team")
	} else if len(item1.children) != 0 {
		t.Fatal("the last node must have no children")
	}

	item1.handler(httptest.NewRecorder(), &http.Request{})
	if !called1 {
		t.Fatal("the handler has not been correctly registered")
	}

	item2, item2Present := dynamic.children["item2"]
	if !item2Present {
		t.Fatal("the last node should be on team")
	} else if len(item2.children) != 0 {
		t.Fatal("the last node must have no children")
	}

	item2.handler(httptest.NewRecorder(), &http.Request{})
	if !called2 {
		t.Fatal("the handler has not been correctly registered")
	}
}

func TestRegisterHandlerWithConflictOnStaticPartOfPath(t *testing.T) {
	// given
	defer func() {
		if r := recover(); r != nil {
			t.Log("successfuly caught the router panic")
		}
	}()
	f1 := func(http.ResponseWriter, *http.Request) {}

	f2 := func(http.ResponseWriter, *http.Request) {}
	// at the same tree level, dymanic path value has to be considered as the same resource
	path1 := []string{"api", "v1", "item1"}
	path2 := []string{"api", "v1", "item1"}
	r := NewDynamicRouter()

	// when
	r.registerHandler(path1, f1)
	r.registerHandler(path2, f2)

	// the router must panic. If not => fatal
	t.Fatal("expect the router to panic")
}

func TestRegisterHandlerWithConflictOnDynamicPartOfPath(t *testing.T) {
	// given
	defer func() {
		if r := recover(); r != nil {
			t.Log("successfuly caught the router panic")
		}
	}()
	f1 := func(http.ResponseWriter, *http.Request) {}

	f2 := func(http.ResponseWriter, *http.Request) {}
	// at the same tree level, dymanic path value has to be considered as the same resource
	path1 := []string{"api", ":someId1", "item1"}
	path2 := []string{"api", ":someId2", "item2"}
	r := NewDynamicRouter()

	// when
	r.registerHandler(path1, f1)
	r.registerHandler(path2, f2)

	// the router must panic. If not => fatal
	t.Fatal("expect the router to panic")
}

func TestFindEndpointOnStaticRoute(t *testing.T) {
	// given
	var called bool
	f := func(http.ResponseWriter, *http.Request) {
		called = true
	}

	req := http.Request{URL: &url.URL{Path: "/api/v1/item/"}}

	r := NewDynamicRouter()

	r.root["api"] = &node{children: make(map[string]*node)}
	r.root["api"].children["v1"] = &node{children: make(map[string]*node)}
	r.root["api"].children["v1"].children["item"] = &node{children: make(map[string]*node), handler: f}

	// when
	n, err := r.findEndpoint(&req)

	if err != nil {
		t.Fatal("expect error to be nil")
	} else if n.handler == nil {
		t.Fatal("expect to hava a non nil handler")
	}
	n.handler(httptest.NewRecorder(), &req)
	if !called {
		t.Fatal("the handler is not the right one")
	}
}

func TestFindEndpointOnDynamicRoute(t *testing.T) {
	// given
	var called bool
	f := func(http.ResponseWriter, *http.Request) {
		called = true
	}

	req := http.Request{URL: &url.URL{Path: "/api/v1/item/12345"}}

	r := NewDynamicRouter()

	r.root["api"] = &node{children: make(map[string]*node)}
	r.root["api"].children["v1"] = &node{children: make(map[string]*node)}
	r.root["api"].children["v1"].children["item"] = &node{children: make(map[string]*node)}
	r.root["api"].children["v1"].children["item"].children[":itemId"] = &node{children: make(map[string]*node), handler: f}

	// when
	n, err := r.findEndpoint(&req)

	if err != nil {
		t.Fatal("expect error to be nil")
	} else if n.handler == nil {
		t.Fatal("expect to hava a non nil handler")
	}
	n.handler(httptest.NewRecorder(), &req)
	if !called {
		t.Fatal("the handler is not the right one")
	}
}

// #######################################################################
// ################## 		Benchmark 		##################
// #######################################################################

func BenchmarkFindEndpointOnStaticRoute(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		f := func(http.ResponseWriter, *http.Request) {}
		r := NewDynamicRouter()
		r.root["api"] = &node{children: make(map[string]*node)}
		r.root["api"].children["v1"] = &node{children: make(map[string]*node)}
		r.root["api"].children["v1"].children["item"] = &node{children: make(map[string]*node), handler: f}
		req := http.Request{URL: &url.URL{Path: "/api/v1/item/"}}
		for pb.Next() {
			r.findEndpoint(&req)
		}
	})
}

func BenchmarkFindEndpointOnDynamicRouteRoute(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		f := func(http.ResponseWriter, *http.Request) {}
		r := NewDynamicRouter()
		r.root["api"] = &node{children: make(map[string]*node)}
		r.root["api"].children["v1"] = &node{children: make(map[string]*node)}
		r.root["api"].children["v1"].children["item"] = &node{children: make(map[string]*node)}
		r.root["api"].children["v1"].children["item"].children[":itemId"] = &node{children: make(map[string]*node), handler: f}
		req := http.Request{URL: &url.URL{Path: "/api/v1/item/12345"}}
		for pb.Next() {
			r.findEndpoint(&req)
		}
	})
}

// test unknown path
// bench router !
// doublon
//
