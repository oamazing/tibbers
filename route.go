package tibbers

type Router interface {
	Group(path string) Router
	Get(path string, handle interface{})
	Put(path string, handle interface{})
	Patch(path string, handle interface{})
	Delete(path string, handle interface{})
	Post(path string, handle interface{})
	Use(...Handle)
}
