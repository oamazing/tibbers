package tibbers

import (
	"fmt"
	"testing"
)

var fakeHandler = func() func(ctx *Context) {
	return func(ctx *Context) {

	}
}

func TestTreeSet(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/hi",
		"/contact",
		"/co",
		"/c",
		"/a",
		"/ab",
		"/doc/",
		"/doc/go_faq",
		"/doc/go1",
		"/α",
		"/β",
	}
	for _, route := range routes {
		tree.addRoute(route, fakeHandler())
	}
}

func TestTreeSetAndGet(t *testing.T) {
	tree := &node{}
	routes := []string{
		"/hi",
		"/contact",
		"/co",
		"/c",
		"/a",
		"/ab",
		"/doc/",
		"/doc/go_faq",
		"/doc/go1",
		"/α",
		"/β",
		"/dsada/dsdsa",
	}
	for _, route := range routes {
		tree.addRoute(route, fakeHandler())
	}
	othersRoutes := []string{
		`/aa`,
		`/dd`,
		`/as/ds`,
	}
	routes = append(routes, othersRoutes...)
	// checkRequest( []string{``})
	for _, path := range routes {
		handler := tree.getValue(path)
		if handler == nil {
			fmt.Printf("%s not found\n", path)
		}
	}
	// Output:
	// /aa not found
	// /dd not found
	// /as/ds not found
}
