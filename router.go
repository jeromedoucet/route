package route

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// FileServerMode allow to adapt some behavior of the file server.
type FileServerMode string

const (
	// classic mode handle all requests considering that there is no routing in front end code.
	// it behave like "classic" web app. When a resource is not found => 404
	Classic FileServerMode = "classic"
	// spa mode considers that all the routing is done in the browser. Some files other than index.html
	// may be loaded, but if a request does not fit a file, the request is changed to serve `/` instead.
	Spa FileServerMode = "spa"
)

// internal representation of a
// routes path segment
type node struct {
	handler  func(context.Context, http.ResponseWriter, *http.Request)
	filters  []HttpFilter
	children map[string]*node
}

type customFileServer struct {
	root    http.Dir
	handler http.Handler
	mode    FileServerMode
}

func containsDotDot(v string) bool {
	if !strings.Contains(v, "..") {
		return false
	}
	for _, ent := range strings.FieldsFunc(v, isSlashRune) {
		if ent == ".." {
			return true
		}
	}
	return false
}

func isSlashRune(r rune) bool {
	return r == '/' || r == '\\'
}

func (fs *customFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if containsDotDot(r.URL.Path) {
		http.Error(w, "URL should not contain '/../' parts", http.StatusBadRequest)
		return
	}
	//if empty, set current directory
	dir := string(fs.root)
	if dir == "" {
		dir = "."
	}

	upath := path.Clean(r.URL.Path)

	//path to file
	name := filepath.FromSlash(path.Join(dir, upath))

	//check if file exists
	f, err := os.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println(fmt.Sprintf("warn: file %s not found for computed path %s on query %s", name, upath, r.URL.Path))
			if fs.mode == Spa {
				r.URL.Path = "/"
			}
		}
	} else {
		defer f.Close()
	}

	fs.handler.ServeHTTP(w, r)

	//http.ServeFile(w, r, name)
}

// DynamicRouter is a simple http router
//
// Implements the http/Handler interface
type DynamicRouter struct {
	root       map[string]*node
	ctx        context.Context
	fileServer *customFileServer
}

// functions that are executed before there corresponding handler.
// if an error is returned, the pipe execution is stopped.
// The filter MUST take care of status and body response.
type HttpFilter func(http.ResponseWriter, *http.Request) bool

// function type used by application code
type Handler func(context.Context, http.ResponseWriter, *http.Request)

type response struct {
	http.ResponseWriter
	http.Hijacker
}

// will wrap the response writer in order
// to controle when the status code will be set in ResponseWriter.
// this is necessary to force 500 status when application
// code do panic, till only one call to WriteHeader is possible.
//
// this is an internal mechanism that should stay hidden
// and must not interfere with application behavior
type responseWrapper struct {
	http.ResponseWriter
	http.Hijacker
	status int
	body   []byte
}

func (w *responseWrapper) WriteHeader(code int) {
	w.status = code
}

func (w *responseWrapper) Write(body []byte) (int, error) {
	w.body = body
	return len(body), nil
}

func (w *responseWrapper) flush() {
	w.ResponseWriter.WriteHeader(w.status)
	w.ResponseWriter.Write(w.body)
}

// NewDynamicRouter create a new DynamicRouter
func NewDynamicRouter() *DynamicRouter {
	r := new(DynamicRouter)
	r.root = make(map[string]*node)
	r.ctx = context.Background()
	return r
}

func (r *DynamicRouter) ServeStaticAt(root string, mode FileServerMode) {
	r.fileServer = &customFileServer{root: http.Dir(root), mode: mode, handler: http.FileServer(http.Dir(root))}
}

// HandleFunc register a new Handler for a given pattern
func (r *DynamicRouter) HandleFunc(pattern string, handler Handler, filters ...HttpFilter) {
	r.registerHandler(SplitPath(pattern), handler, filters...)
}

// http/Handler implementation
func (r *DynamicRouter) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	w := &responseWrapper{ResponseWriter: res, status: 200, body: []byte{}}
	hj, ok := res.(http.Hijacker)
	if ok {
		w.Hijacker = hj
	}
	defer func() {
		if r := recover(); r != nil {
			// we dunno what's happened so, we set the
			// status code to 500
			w.WriteHeader(http.StatusInternalServerError)
			w.flush()
		}
	}()
	n, err := r.findEndpoint(req)
	if err != nil {
		if r.fileServer == nil {
			w.WriteHeader(http.StatusNotFound)
			w.flush()
		} else {
			r.fileServer.ServeHTTP(res, req)
		}
	} else if n.handler != nil {
		// we pass all filter in the right order. if one return false
		// we return, assuming that everything has been written in response
		for _, filter := range n.filters {
			if !filter(w, req) {
				w.flush()
				return
			}
		}
		n.handler(r.ctx, w, req)
		w.flush()
	}
}

func (r *DynamicRouter) registerHandler(paths []string, handler Handler, filters ...HttpFilter) {
	if handler == nil {
		panic("handler cannot be nil")
	} else if len(paths) < 1 {
		panic("path cannot be nil")
	}
	children := r.root
	var n *node
	var ok bool
	for _, path := range paths {
		if path == "" {
			continue
		}
		/*
		 * we only consider static and dynamic identifier of the path.
		 *
		 * For static :
		 * If at a given non terminal node, the resource
		 * already exist, and if the identifier is static, we just
		 * pass to the next level.
		 *
		 * For dynamic :
		 * if the identifier of the resource is dynamic and if a
		 * dynamic identifier already exist with another name, the router will panic.
		 *
		 * Common :
		 * If the node denoted by the incoming path already has a handler, the router will panic
		 */
		if strings.HasPrefix(path, ":") {
			for m := range children {
				if strings.HasPrefix(m, ":") && path != m {
					panic("a dynamic identifier has already been registered at that level")
				}
			}
		}
		n, ok = children[path]
		if !ok {
			n = &node{}
			n.children = make(map[string]*node)
			children[path] = n
		}
		children = n.children
	}
	if n.handler != nil {
		panic("a handler is already registered for this path")
	}
	n.handler = handler
	n.filters = filters
}

func (r *DynamicRouter) findEndpoint(req *http.Request) (n *node, err error) {
	// todo clean path
	// todo check url encoder
	return parseTree(r.root, SplitPath(req.URL.Path))
}

// SplitPath is an utils function that will
// split the path of a request.
func SplitPath(path string) []string {
	p := strings.TrimPrefix(path, "/")
	return strings.Split(strings.TrimSuffix(p, "/"), "/")
}

func parseTree(children map[string]*node, path []string) (*node, error) {
	n, ok := children[path[0]]
	if !ok {
		// if no static path found, look for a dynamic one
		// todo make some optimization
		for p, dn := range children {
			if strings.HasPrefix(p, ":") {
				n = dn
				break
			}
		}
		if n == nil {
			return n, errors.New("unknown path")
		}

	}
	if len(path) > 1 {
		return parseTree(n.children, path[1:])
	}
	return n, nil
}
