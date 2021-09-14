package tibbers

import (
	"log"
	"net/http"
	"reflect"
	"sync"
)

type Handle func(*Context)

var notFound = []byte("Not Found")

type Tibbers struct {
	routes   map[string]*node
	pool     sync.Pool
	basePath string
	handles  []Handle
	notFound []byte
}

func New() *Tibbers {
	tibbers := &Tibbers{
		routes:   make(map[string]*node),
		notFound: notFound,
	}
	tibbers.pool.New = func() interface{} {
		return &Context{tibbers: tibbers}
	}
	return tibbers
}

func (tibbers *Tibbers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := tibbers.pool.Get().(*Context)
	ctx.reset()
	ctx.Request = r
	ctx.Writer = w
	if tree, ok := tibbers.routes[r.Method]; ok {
		handles := tree.getValue(r.URL.Path)
		if len(handles) > 0 {
			ctx.handles = handles
			ctx.Next()
		} else {
			ctx.Writer.WriteHeader(http.StatusNotFound)
			ctx.Writer.Write(tibbers.notFound)
		}
	} else {
		ctx.Writer.WriteHeader(http.StatusNotFound)
		ctx.Writer.Write(tibbers.notFound)
	}
	tibbers.pool.Put(ctx)
}

func (tibbers *Tibbers) Run(addr string) error {
	return http.ListenAndServe(addr, tibbers)
}

func (tibbers *Tibbers) GET(path string, handle interface{}) {
	tibbers.Handle(http.MethodGet, path, convertHandle(handle))
}

func (tibbers *Tibbers) POST(path string, handle interface{}) {
	tibbers.Handle(http.MethodPost, path, convertHandle(handle))
}

func (tibbers *Tibbers) PUT(path string, handle interface{}) {
	tibbers.Handle(http.MethodPut, path, convertHandle(handle))
}
func (tibbers *Tibbers) PATCH(path string, handle interface{}) {
	tibbers.Handle(http.MethodPatch, path, convertHandle(handle))
}
func (tibbers *Tibbers) DELETE(path string, handle interface{}) {
	tibbers.Handle(http.MethodDelete, path, convertHandle(handle))
}
func (tibbers *Tibbers) Group(path string) Route {
	return &Tibbers{
		basePath: tibbers.basePath + path,
		routes:   tibbers.routes,
		handles:  tibbers.handles,
	}
}
func (tibbers *Tibbers) Use(handles ...Handle) {
	tibbers.handles = append(tibbers.handles, handles...)
}

func (tibbers *Tibbers) Handle(method, path string, handle Handle) {
	_, ok := tibbers.routes[method]
	if !ok {
		tibbers.routes[method] = &node{}
	}
	tree := tibbers.routes[method]
	log.Printf("add route %s", tibbers.basePath+path)
	tree.addRoute(tibbers.basePath+path, append(tibbers.handles, handle)...)
}

func convertHandle(h interface{}) Handle {
	if handler, ok := h.(func(*Context)); ok {
		return handler
	}
	val := reflect.ValueOf(h)
	typ := val.Type()
	if typ.Kind() != reflect.Func {
		log.Panic("handler must be a func")
	}
	if typ.NumIn() != 2 {
		log.Panic("handler func must have exactly two parameters.")
	}
	if typ.NumOut() != 0 {
		log.Panic("handler func must have no return values.")
	}
	reqConvertFunc := newReqConvertFunc(typ.In(0))
	respType := newRespConvertFunc(typ.In(1))
	return func(ctx *Context) {
		req, err := reqConvertFunc(ctx)
		if err != nil {
			log.Panic(err)
		}
		resp := reflect.New(respType)
		val.Call([]reflect.Value{req, resp})
		var data interface{}
		Traverse(resp.Elem(), func(val reflect.Value, field reflect.StructField) bool {
			switch field.Name {
			case `Error`:
				if e := val.Interface(); err != nil {
					err = e.(error)
				}
			case `Data`:
				data = val.Interface()
			}
			return true
		})
		ctx.Data(data, err)
	}
}

func (tibbers *Tibbers) SetNotFound(data []byte) {
	tibbers.notFound = data
}
