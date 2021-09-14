package tibbers

import (
	"encoding/json"
	"net/http"
	"reflect"
	"sync"
	"time"
)

const abort = 1>>15 - 1

type Context struct {
	Request *http.Request
	Writer  http.ResponseWriter
	tibbers *Tibbers
	index   int8
	handles []Handle
	mu      sync.RWMutex
	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[string]interface{}
}

func (ctx *Context) reset() {
	ctx.handles = nil
	ctx.index = -1
}

func (ctx *Context) Next() {
	ctx.index++
	for ctx.index < int8(len(ctx.handles)) {
		ctx.handles[ctx.index](ctx)
		ctx.index++
	}
}

func (ctx *Context) Data(data interface{}, err error) {
	body := struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data,omitempty"`
	}{}
	if err == nil {
		body.Code = 0
		body.Message = `success`
	} else {
		body.Code = -1
		body.Message = err.Error()
	}
	if err == nil && data != nil && !isNilValue(data) {
		body.Data = data
	}
	ctx.Json(http.StatusOK, body)
}

func (ctx *Context) Json(statusCode int, data interface{}) {
	ctx.Writer.Header().Set(`Content-Type`, `application/json; charset=utf-8`)
	ctx.Writer.WriteHeader(statusCode)

	encoder := json.NewEncoder(ctx)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(data); err != nil {
		// ctx.SetError(err)
		ctx.Write([]byte(`{"code":"json-marshal-error","message":"json marshal error"}`))
	}
}

func isNilValue(itfc interface{}) bool {
	v := reflect.ValueOf(itfc)
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface:
		return v.IsNil()
	}
	return false
}

func (ctx *Context) Write(content []byte) (int, error) {
	return ctx.Writer.Write(content)
}

// Set is used to store a new key/value pair exclusively for this context.
// It also lazy initializes  c.Keys if it was not used previously.
func (ctx *Context) Set(key string, value interface{}) {
	ctx.mu.Lock()
	if ctx.Keys == nil {
		ctx.Keys = make(map[string]interface{})
	}

	ctx.Keys[key] = value
	ctx.mu.Unlock()
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (ctx *Context) Get(key string) (value interface{}, exists bool) {
	ctx.mu.RLock()
	value, exists = ctx.Keys[key]
	ctx.mu.RUnlock()
	return
}

// GetString returns the value associated with the key as a string.
func (c *Context) GetString(key string) (s string) {
	if val, ok := c.Get(key); ok && val != nil {
		s, _ = val.(string)
	}
	return
}

// GetBool returns the value associated with the key as a boolean.
func (c *Context) GetBool(key string) (b bool) {
	if val, ok := c.Get(key); ok && val != nil {
		b, _ = val.(bool)
	}
	return
}

// GetInt returns the value associated with the key as an integer.
func (c *Context) GetInt(key string) (i int) {
	if val, ok := c.Get(key); ok && val != nil {
		i, _ = val.(int)
	}
	return
}

// GetInt64 returns the value associated with the key as an integer.
func (c *Context) GetInt64(key string) (i64 int64) {
	if val, ok := c.Get(key); ok && val != nil {
		i64, _ = val.(int64)
	}
	return
}

// GetUint returns the value associated with the key as an unsigned integer.
func (c *Context) GetUint(key string) (ui uint) {
	if val, ok := c.Get(key); ok && val != nil {
		ui, _ = val.(uint)
	}
	return
}

// GetUint64 returns the value associated with the key as an unsigned integer.
func (c *Context) GetUint64(key string) (ui64 uint64) {
	if val, ok := c.Get(key); ok && val != nil {
		ui64, _ = val.(uint64)
	}
	return
}

/************************************/
/***** GOLANG.ORG/X/NET/CONTEXT *****/
/************************************/

// Deadline returns that there is no deadline (ok==false) when c.Request has no Context.
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	if c.Request == nil || c.Request.Context() == nil {
		return
	}
	return c.Request.Context().Deadline()
}

// Done returns nil (chan which will wait forever) when c.Request has no Context.
func (c *Context) Done() <-chan struct{} {
	if c.Request == nil || c.Request.Context() == nil {
		return nil
	}
	return c.Request.Context().Done()
}

// Err returns nil when c.Request has no Context.
func (c *Context) Err() error {
	if c.Request == nil || c.Request.Context() == nil {
		return nil
	}
	return c.Request.Context().Err()
}

// Value returns the value associated with this context for key, or nil
// if no value is associated with key. Successive calls to Value with
// the same key returns the same result.
func (c *Context) Value(key interface{}) interface{} {
	if key == 0 {
		return c.Request
	}
	if keyAsString, ok := key.(string); ok {
		if val, exists := c.Get(keyAsString); exists {
			return val
		}
	}
	if c.Request == nil || c.Request.Context() == nil {
		return nil
	}
	return c.Request.Context().Value(key)
}
