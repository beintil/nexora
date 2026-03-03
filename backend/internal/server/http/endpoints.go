package http

import (
	"sort"

	"github.com/gorilla/mux"
)

// Endpoint describes a registered HTTP endpoint (method + path).
type Endpoint struct {
	Method string
	Path   string
}

// RegisteredEndpoints walks the router and returns all registered route (method, path) pairs.
// Routes with no methods (e.g. path-only matchers) are skipped.
// Result is sorted by path, then by method.
func RegisteredEndpoints(router *mux.Router) []Endpoint {
	var list []Endpoint
	router.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil || path == "" {
			return nil
		}
		methods, err := route.GetMethods()
		if err != nil || len(methods) == 0 {
			return nil
		}
		for _, m := range methods {
			list = append(list, Endpoint{Method: m, Path: path})
		}
		return nil
	})
	sort.Slice(list, func(i, j int) bool {
		if list[i].Path != list[j].Path {
			return list[i].Path < list[j].Path
		}
		return list[i].Method < list[j].Method
	})
	return list
}
