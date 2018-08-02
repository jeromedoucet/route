package route_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeromedoucet/route"
)

func TestDynamicRoute200(t *testing.T) {
	// given
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}
	router := route.NewDynamicRouter()
	router.HandleFunc("/tests/:testId", handler)
	s := httptest.NewServer(router)
	defer s.Close()

	resp, err := http.Get(fmt.Sprintf("%s/tests/1", s.URL))

	// then
	if err != nil {
		t.Fatalf("Expect to have no error, but got %s", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Expect 200 return code.Got %d", resp.StatusCode)
	}
}

func TestDynamicRouteFilter(t *testing.T) {
	// given
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	filter := func(w http.ResponseWriter, r *http.Request) bool {
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}
	router := route.NewDynamicRouter()
	router.HandleFunc("/tests/:testId", handler, filter)
	s := httptest.NewServer(router)
	defer s.Close()

	resp, err := http.Get(fmt.Sprintf("%s/tests/1", s.URL))

	// then
	if err != nil {
		t.Fatalf("Expect to have no error, but got %s", err.Error())
	}

	if resp.StatusCode != 401 {
		t.Fatalf("Expect 401 return code.Got %d", resp.StatusCode)
	}
}

func TestDynamicRoutePanic(t *testing.T) {
	// given
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		panic(errors.New("something really bad"))
	}
	router := route.NewDynamicRouter()
	router.HandleFunc("/tests/:testId", handler)
	s := httptest.NewServer(router)
	defer s.Close()

	resp, err := http.Get(fmt.Sprintf("%s/tests/1", s.URL))

	// then
	if err != nil {
		t.Fatalf("Expect to have no error, but got %s", err.Error())
	}

	if resp.StatusCode != 500 {
		t.Fatalf("Expect 500 return code.Got %d", resp.StatusCode)
	}
}
