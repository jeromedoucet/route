package route_test

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/squarescale/route"
)

func TestHijack(t *testing.T) {
	// given
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		_, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}
	router := route.NewDynamicRouter()
	router.HandleFunc("/tests", handler)
	s := httptest.NewServer(router)
	defer s.Close()

	resp, err := http.Get(fmt.Sprintf("%s/tests", s.URL))

	// then
	if err != nil {
		t.Fatalf("Expect to have no error, but got %s", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Expect 200 return code.Got %d", resp.StatusCode)
	}
}

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

func TestServeStaticClassique(t *testing.T) {
	// given
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}
	router := route.NewDynamicRouter()
	router.HandleFunc("/tests/:testId", handler)
	router.ServeStaticAt("fixtures/", route.Classic)
	s := httptest.NewServer(router)
	defer s.Close()

	resp, err := http.Get(fmt.Sprintf("%s/", s.URL))

	// then
	if err != nil {
		t.Fatalf("Expect to have no error, but got %s", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Expect 200 return code.Got %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	payloadResp, _ := ioutil.ReadAll(resp.Body)

	expectedContent := `<!DOCTYPE html><html lang="en"></html>`

	if strings.Trim(string(payloadResp), "\n") != expectedContent {
		t.Fatalf("expect %s, but got %s", expectedContent, string(payloadResp))
	}
}

func TestServeStaticClassique404(t *testing.T) {
	// given
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}
	router := route.NewDynamicRouter()
	router.HandleFunc("/tests/:testId", handler)
	router.ServeStaticAt("fixtures/", route.Classic)
	s := httptest.NewServer(router)
	defer s.Close()

	resp, err := http.Get(fmt.Sprintf("%s/toto/titi.html", s.URL))

	// then
	if err != nil {
		t.Fatalf("Expect to have no error, but got %s", err.Error())
	}

	if resp.StatusCode != 404 {
		t.Fatalf("Expect 404 return code.Got %d", resp.StatusCode)
	}
}

func TestServeStaticSpa(t *testing.T) {
	// given
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}
	router := route.NewDynamicRouter()
	router.HandleFunc("/tests/:testId", handler)
	router.ServeStaticAt("fixtures/", route.Spa)
	s := httptest.NewServer(router)
	defer s.Close()

	resp, err := http.Get(fmt.Sprintf("%s/", s.URL))

	// then
	if err != nil {
		t.Fatalf("Expect to have no error, but got %s", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Expect 200 return code.Got %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	payloadResp, _ := ioutil.ReadAll(resp.Body)

	expectedContent := `<!DOCTYPE html><html lang="en"></html>`

	if strings.Trim(string(payloadResp), "\n") != expectedContent {
		t.Fatalf("expect %s, but got %s", expectedContent, string(payloadResp))
	}
}

func TestServeStaticSpaRedirectWhenNotFound(t *testing.T) {
	// given
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}
	router := route.NewDynamicRouter()
	router.HandleFunc("/tests/:testId", handler)
	router.ServeStaticAt("fixtures/", route.Spa)
	s := httptest.NewServer(router)
	defer s.Close()

	resp, err := http.Get(fmt.Sprintf("%s/toto/titi", s.URL))

	// then
	if err != nil {
		t.Fatalf("Expect to have no error, but got %s", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Expect 200 return code.Got %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	payloadResp, _ := ioutil.ReadAll(resp.Body)

	expectedContent := `<!DOCTYPE html><html lang="en"></html>`

	if strings.Trim(string(payloadResp), "\n") != expectedContent {
		t.Fatalf("expect %s, but got %s", expectedContent, string(payloadResp))
	}
}

func TestServeStaticDotProtection(t *testing.T) {
	// given
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}
	router := route.NewDynamicRouter()
	router.HandleFunc("/tests/:testId", handler)
	router.ServeStaticAt("fixtures/", route.Classic)
	s := httptest.NewServer(router)
	defer s.Close()

	resp, err := http.Get(fmt.Sprintf("%s/../some-protected-resources", s.URL))

	// then
	if err != nil {
		t.Fatalf("Expect to have no error, but got %s", err.Error())
	}

	if resp.StatusCode != 400 {
		t.Fatalf("Expect 400 return code.Got %d", resp.StatusCode)
	}
}
