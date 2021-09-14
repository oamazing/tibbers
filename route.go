package tibbers

type Route interface {
	Group(path string) Route
	GET(path string, handle interface{})
	PUT(path string, handle interface{})
	PATCH(path string, handle interface{})
	DELETE(path string, handle interface{})
	POST(path string, handle interface{})
	Use(...Handle)
}
