package main

import (
	"log"

	"./blog"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func main() {
	router := router.New()
	blog := blog.Blog{}

	blog.Path = "/blog/"
	blog.Name = "LowkeyCoding's Blog"
	blog.Description = "This is my personal blog for random things i find interesting"
	blog.PostsPerPage = 10
	blog.JWTSecret = "VerySecretSecret"

	err := blog.Setup(router, "file:./database.db")
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(fasthttp.ListenAndServe("0.0.0.0:8080", router.Handler))
}
