package main

import (
	"fmt"
	"log"

	"./blog"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func index(ctx *fasthttp.RequestCtx) {
	ctx.SendFile("./index.html")
}

func styles(ctx *fasthttp.RequestCtx) {
	path := fmt.Sprintf("./styles/%s", ctx.UserValue("stylesheet"))
	ctx.SendFile(path)
}

func main() {
	router := router.New()
	blog := blog.Blog{}

	router.GET("/", index)
	router.GET("/styles/{stylesheet}", styles)

	blog.Path = "/blog/"
	blog.Name = "LowkeyCoding's Blog"
	blog.Description = "This is my personal blog for random things i find interesting"
	blog.PostsPerPage = 10
	blog.JWTSecret = "Spiasdandoiaubnfdaios"

	err := blog.Setup(router, "file:./database.db")
	if err != nil {
		log.Fatal(err)
	}
	err = blog.CreateUser("LowkeyCoding", "b109f3bbbc244eb82441917ed06d618b9008dd09b3befd1b5e07394c706a8bb980b1d7785e5976ec049b46df5f1326af5a2ea6d103fd07c95385ffab0cacbc86")
	if err != nil {
		log.Println(err)
	}
	log.Fatal(fasthttp.ListenAndServe("0.0.0.0:8080", router.Handler))
}
