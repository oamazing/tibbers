package tibbers

import (
	"encoding/json"
	"io/ioutil"
	"reflect"

	"github.com/oamazing/tibbers/convert"
)

func newReqConvertFunc(typ reflect.Type) func(*Context) (reflect.Value, error) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	var err error
	return func(ctx *Context) (reflect.Value, error) {
		ptr := reflect.New(typ)
		req := ptr.Elem()
		convert.Traverse(req, func(val reflect.Value, field reflect.StructField) bool {
			switch field.Name {
			case "Query":
				convertNilPtr(val)
				err = convert.Query(val, ctx.Request.URL.Query())
			case `Body`:
				err = convertReqBody(val, ctx)
			}
			return false
		})
		if err != nil {
			return reflect.Value{}, err
		}
		return req, nil
	}
}

func convertNilPtr(v reflect.Value) {
	if v.Kind() == reflect.Ptr && v.IsNil() && v.CanSet() {
		v.Set(reflect.New(v.Type().Elem()))
	}
}

func convertReqBody(value reflect.Value, ctx *Context) error {

	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, value.Addr().Interface())
}
