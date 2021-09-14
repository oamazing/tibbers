package tibbers

import (
	"log"
	"reflect"
)

func newRespConvertFunc(typ reflect.Type) reflect.Type {
	if typ.Kind() != reflect.Ptr {
		log.Panic("resp parameter of handler func must be a struct pointer.")
	}
	typ = typ.Elem()
	if typ.Kind() != reflect.Struct {
		log.Panic("resp parameter of handler func must be a struct pointer.")
	}
	return typ
}
