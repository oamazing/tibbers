package tibbers

import (
	"errors"
	"log"
	"net/http"
	"reflect"
	"sync"

	"github.com/oamazing/tibbers/convert"
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
	tibbers.handleHttp(ctx)
	tibbers.pool.Put(ctx)
}

func (tibbers *Tibbers) handleHttp(ctx *Context) {
	defer func() {
		if err := recover(); err != nil {
			ctx.Data(nil, errors.New(err.(string)))
		}
	}()
	if tree, ok := tibbers.routes[ctx.Request.Method]; ok {
		handles := tree.getValue(ctx.Request.URL.Path)
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
}

func (tibbers *Tibbers) Run(addr string) error {
	return http.ListenAndServe(addr, tibbers)
}

func (tibbers *Tibbers) Get(path string, handle interface{}) {
	tibbers.Handle(http.MethodGet, path, convertHandle(handle))
}

func (tibbers *Tibbers) Post(path string, handle interface{}) {
	tibbers.Handle(http.MethodPost, path, convertHandle(handle))
}

func (tibbers *Tibbers) Put(path string, handle interface{}) {
	tibbers.Handle(http.MethodPut, path, convertHandle(handle))
}
func (tibbers *Tibbers) Patch(path string, handle interface{}) {
	tibbers.Handle(http.MethodPatch, path, convertHandle(handle))
}
func (tibbers *Tibbers) Delete(path string, handle interface{}) {
	tibbers.Handle(http.MethodDelete, path, convertHandle(handle))
}
func (tibbers *Tibbers) Group(path string) Router {
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
		convert.Traverse(resp.Elem(), func(val reflect.Value, field reflect.StructField) bool {
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
