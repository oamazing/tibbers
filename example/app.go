package main

import (
	"fmt"

	"github.com/oamazing/tibbers"
)

type Resp struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func main() {
	app := tibbers.New()
	app.GET(`/test`, func(
		req struct {
			Query struct {
				Id   int64  `json:"id"`
				Name string `json:"name"`
			}
		}, resp *struct {
			Error error
			Data  *Resp
		}) {
		fmt.Println("test")
		resp.Data = &Resp{
			Id:   req.Query.Id,
			Name: req.Query.Name,
		}
	})
	if err := app.Run(":8080"); err != nil {
		panic(err)
	}
}
