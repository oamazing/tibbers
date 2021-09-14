package tibbers

import "reflect"

func newReqConvertFunc(typ reflect.Type) func(*Context) (reflect.Value, error) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return func(ctx *Context) (reflect.Value, error) {
		ptr := reflect.New(typ)
		req := ptr.Elem()
		Traverse(req, func(val reflect.Value, field reflect.StructField) bool {
			switch field.Name {
			case "Query":
			case `Body`:
			}
			return false
		})
		return req, nil
	}
}

// func convertNilPtr(v reflect.Value) {
// 	if v.Kind() == reflect.Ptr && v.IsNil() && v.CanSet() {
// 		v.Set(reflect.New(v.Type().Elem()))
// 	}
// }
