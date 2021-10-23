package convert

import "reflect"

func Traverse(val reflect.Value, fn func(val reflect.Value, field reflect.StructField) bool) bool {
	typ := val.Type()
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.FieldByIndex(field.Index)
		if field.Anonymous {
			fieldTyp := field.Type
			if fieldTyp.Kind() == reflect.Ptr && fieldTyp.Elem().Kind() == reflect.Struct {
				fieldTyp = fieldTyp.Elem()
				if fieldVal.IsNil() {
					if fieldVal.CanSet() {
						fieldVal.Set(reflect.New(fieldTyp))
					} else {
						continue
					}
				}
				fieldVal = fieldVal.Elem()
			}
			if fieldTyp.Kind() == reflect.Struct {
				if !Traverse(fieldVal, fn) {
					return false // stop traverse
				}
				continue
			}
		}
		if field.Name[0] >= 'A' && field.Name[0] <= 'Z' {
			if !fn(fieldVal, field) {
				return false
			}
		}
	}
	return true
}
