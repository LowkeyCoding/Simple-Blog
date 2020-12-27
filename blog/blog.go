package blog

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/gobuffalo/velvet"

	// sqlite3 driver
	_ "github.com/mattn/go-sqlite3"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

// Blog struct
type Blog struct {
	Path         string
	Name         string
	Description  string
	Database     *sql.DB
	Posts        []Post
	PostsPerPage int
	JWTSecret    string
}

// User struct
type User struct {
	ID       int
	Username string
	Password string
}

type credentials struct {
	Username []byte
	Password []byte
	jwt.StandardClaims
}

// Post struct
type Post struct {
	ID          int
	Title       string
	Description string
	Content     string
	HTML        string
	Time        string
}

// Response struct
type Response struct {
	Errors   []string
	Response string
}

/* Helper functions */
func (blog *Blog) getPostCount() int {
	var postCount *int
	row := blog.Database.QueryRow("SELECT COUNT(*) FROM BlogPosts")
	row.Scan(&postCount)
	return *postCount
}

func (blog *Blog) getUserCount() int {
	var userCount *int
	row := blog.Database.QueryRow("SELECT COUNT(*) FROM Users")
	row.Scan(&userCount)
	return *userCount
}

func (response *Response) appendError(err error) {
	if err != nil {
		response.Errors = append(response.Errors, err.Error())
	}
}

func (response *Response) toJSON() (string, error) {
	json, err := json.Marshal(&response)
	return string(json), err
}

func (response *Response) unkownError() string {
	return "Response: {\"Errors\":\"Unkown error please try again.\",\"Response\":\"failed\"}"
}

func (response *Response) writeResponse(arg interface{}) {
	responseString, err := response.toJSON()
	switch ctx := (arg).(type) {
	case *fasthttp.RequestCtx:
		if err != nil {
			ctx.WriteString(response.unkownError())
		} else {
			ctx.WriteString(responseString)
		}
	case *velvet.Context:
		ctx.Set("Response", response)
	}

}

func (response *Response) setResponse(_response string) {
	response.Response = _response
}

func handleError(err error) {
	if err != nil {
		log.Println(err)
	}
}

/* Authentication middleware */
func (blog *Blog) createJWTToken(username []byte, password []byte) (string, time.Time) {
	expires := time.Now().Add(30 * time.Minute)
	claims := &credentials{
		Username: username,
		Password: password,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expires.Unix(),
		},
	}
	tokenObject := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tokenString, err := tokenObject.SignedString([]byte("SuperSecretKey"))
	if err != nil {
		log.Print(err)
	}
	return tokenString, expires
}

func (blog *Blog) validateJWTToken(tokenString string) (*jwt.Token, *credentials, error) {
	credentials := &credentials{}
	tokenObject, err := jwt.ParseWithClaims(tokenString, credentials, func(token *jwt.Token) (interface{}, error) {
		return []byte("SuperSecretKey"), nil
	})
	return tokenObject, credentials, err
}

func (blog *Blog) validateJWTTokenMiddleware(ctx *fasthttp.RequestCtx) error {
	tokenByte := ctx.Request.Header.Cookie("JWT")

	if len(tokenByte) != 0 {
		token, credentials, err := blog.validateJWTToken(string(tokenByte))

		user := blog.getUserByUsername(string(credentials.Username))

		if user.Password != string(credentials.Password) {
			return errors.New("Invalid password")
		} else if err != nil {
			return errors.New("Could not validate JWT token")
		} else if !token.Valid {
			return errors.New("your session is expired, login again please")
		}
	} else {
		return errors.New("Login required")
	}
	return nil
}

/* DATABASE QUERIES */
func (blog *Blog) initialiseDatabase(database string) error {
	var err error
	blog.Database, err = sql.Open("sqlite3", database)

	if err != nil {
		return err
	}
	// Protects database by only allowing a single connection to the database.
	blog.Database.SetMaxOpenConns(1)

	// Setup database tables.
	statement, err := blog.Database.Prepare(`
		CREATE TABLE IF NOT EXISTS Users(
			ID INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			Username TEXT NOT NULL,
			Password TEXT NOT NULL
		);
	`)
	if err != nil {
		return err
	}

	statement.Exec()

	statement, err = blog.Database.Prepare(`
		CREATE TABLE IF NOT EXISTS BlogPosts(
			ID INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			Title TEXT NOT NULL,
			Description TEXT NOT NULL,
			Content TEXT NOT NULL,
			HTML TEXT NOT NULL,
			Time TEXT NOT NULL
		);
	`)

	if err != nil {
		return err
	}

	statement.Exec()

	return nil
}

func (blog *Blog) createUser(username string, password string) error {
	var err error
	checkUser := blog.getUserByUsername(username)
	if checkUser.Username != username {
		statement, _ := blog.Database.Prepare("INSERT INTO Users(Username, Password) VALUES (?, ?)")
		_, err = statement.Exec(username, password)
	}
	return err
}

func (blog *Blog) getUserByUsername(username string) User {
	result := blog.Database.QueryRow("SELECT * FROM Users WHERE Username=$1", username)

	user := User{}

	result.Scan(&user.ID, &user.Username, &user.Password)

	return user
}

func (blog *Blog) createPost(post *Post) (int64, error) {
	var id int64
	id = 0
	ctime := time.Now().Format(time.RFC3339)
	statement, _ := blog.Database.Prepare("INSERT INTO BlogPosts(Title, Description, Content, HTML, Time) VALUES (?, ?, ?, ?, ?)")
	res, err := statement.Exec(post.Title, post.Description, post.Content, post.HTML, ctime)
	id, err = res.LastInsertId()
	return id, err
}

func (blog *Blog) updatePost(post *Post) error {
	ctime := time.Now().Format(time.RFC3339)
	statement, err := blog.Database.Prepare("UPDATE BlogPosts SET Title=?, Description=?, Content=?, HTML=?, Time=? WHERE ID=?")
	if err != nil {
		return err
	}
	_, err = statement.Exec(post.Title, post.Description, post.Content, post.HTML, ctime, post.ID)
	if err != nil {
		return err
	}
	return nil
}

func (blog *Blog) deletePost(id int) error {
	statement, err := blog.Database.Prepare("DELETE FROM BlogPosts WHERE ID=?")
	if err != nil {
		return err
	}
	_, err = statement.Exec(id)
	if err != nil {
		return err
	}
	return nil
}

func (blog *Blog) getPostsFromPageNumber(page int) error {
	page--
	result, err := blog.Database.Query("SELECT * FROM BlogPosts LIMIT $1 OFFSET $2",
		blog.PostsPerPage, blog.PostsPerPage*page)

	if err != nil {
		return err
	}

	Posts := []Post{}

	for result.Next() {
		post := Post{}
		err := result.Scan(&post.ID, &post.Title, &post.Description, &post.HTML, &post.Content, &post.Time)
		if err != nil {
			return err
		}
		Posts = append(Posts, post)
	}
	blog.Posts = Posts
	return nil
}

func (blog *Blog) getPostFromID(ID int) Post {
	result := blog.Database.QueryRow("SELECT * FROM BlogPosts WHERE ID=$1", ID)
	post := Post{}
	result.Scan(&post.ID, &post.Title, &post.Description, &post.Content, &post.HTML, &post.Time)
	return post
}

/* GET ROUTES */
func (blog *Blog) indexRoute(ctx *fasthttp.RequestCtx) {
	var response Response
	var err error
	var pageCount []int
	hbx := velvet.NewContext()

	_pagenumber, _ := ctx.UserValue("pagenumber").(string)

	pagenumber := 1

	if _pagenumber != "" {
		pagenumber, err = strconv.Atoi(_pagenumber)
	}

	response.appendError(err)

	if pagenumber < 1 {
		pagenumber = 1
	}

	err = blog.getPostsFromPageNumber(pagenumber)
	response.appendError(err)

	err = blog.validateJWTTokenMiddleware(ctx)

	if err != nil {
		if err.Error() != "Login required" {
			response.appendError(err)
		}
		hbx.Set("LoggedIn", false)
	} else {
		hbx.Set("LoggedIn", true)
	}

	maxPageCount := blog.getPostCount() / blog.PostsPerPage

	if pagenumber <= maxPageCount {
		offsetNegative := 2
		offsetPositive := 2
		for i := pagenumber - offsetNegative; i <= pagenumber; i++ {
			if i >= 1 {
				pageCount = append(pageCount, i)
			} else {
				offsetPositive++
			}
		}
		for i := pagenumber + 1; i <= pagenumber+offsetPositive; i++ {
			if i > 1 && i <= maxPageCount {
				pageCount = append(pageCount, i)
			} else if pagenumber-offsetNegative-1 > 1 {
				offsetNegative++
				pageCount = append(pageCount, pagenumber-offsetNegative)
			}
		}
		sort.Ints(pageCount)
		if pagenumber > 3 {
			hbx.Set("MinPageCount", 1)
		}
		hbx.Set("PageCount", pageCount)
		if pagenumber+2 < maxPageCount {
			hbx.Set("MaxPageCount", maxPageCount)
		}
	}
	if pagenumber < maxPageCount {
		hbx.Set("NextPage", pagenumber+1)
	}
	if pagenumber > 0 {
		hbx.Set("LastPage", pagenumber-1)
	}

	hbx.Set("CurrentPage", pagenumber)
	hbx.Set("Blog", blog)

	file, err := ioutil.ReadFile("./blog/views/index.handlebars")
	handleError(err)

	response.writeResponse(hbx)
	result, err := velvet.Render(string(file), hbx)
	handleError(err)

	ctx.SetContentType("text/html")
	ctx.WriteString(result)
}

func (blog *Blog) postRoute(ctx *fasthttp.RequestCtx) {
	var response Response
	var err error
	var count *int
	hbx := velvet.NewContext()
	_ID, _ := ctx.UserValue("ID").(string)
	row := blog.Database.QueryRow("SELECT COUNT(*) FROM BlogPosts")
	row.Scan(&count)
	ID := 0

	if _ID != "" {
		ID, err = strconv.Atoi(_ID)
		response.appendError(err)
	}

	post := blog.getPostFromID(ID)
	hbx.Set("Post", post)
	hbx.Set("Blog", blog)

	err = blog.validateJWTTokenMiddleware(ctx)
	if err != nil {
		hbx.Set("LoggedIn", false)
	} else {
		hbx.Set("LoggedIn", true)
	}

	file, err := ioutil.ReadFile("./blog/views/post.handlebars")
	handleError(err)

	response.writeResponse(hbx)
	result, err := velvet.Render(string(file), hbx)
	handleError(err)

	ctx.SetContentType("text/html")
	ctx.WriteString(result)
}

func (blog *Blog) editorRoute(ctx *fasthttp.RequestCtx) {
	var response Response
	var count *int

	hbx := velvet.NewContext()
	err := blog.validateJWTTokenMiddleware(ctx)
	response.appendError(err)

	if err != nil {
		ctx.Redirect("/blog/", fasthttp.StatusTemporaryRedirect)
	} else {
		row := blog.Database.QueryRow("SELECT COUNT(*) FROM BlogPosts")
		row.Scan(&count)

		_ID, _ := ctx.UserValue("ID").(string)

		ID := 0

		if _ID != "" {
			ID, err = strconv.Atoi(_ID)
			response.appendError(err)
		}

		if ID != 0 {
			post := blog.getPostFromID(ID)
			hbx.Set("Action", "updatePost")
			hbx.Set("Post", post)
		} else {
			hbx.Set("Action", "publishPost")
		}

		hbx.Set("Blog", blog)

		file, err := ioutil.ReadFile("./blog/views/editor.handlebars")
		handleError(err)

		response.writeResponse(hbx)
		result, err := velvet.Render(string(file), hbx)

		ctx.SetContentType("text/html")
		ctx.WriteString(result)
	}
}

func (blog *Blog) loginViewRoute(ctx *fasthttp.RequestCtx) {
	var response Response

	hbx := velvet.NewContext()

	if string(ctx.Path()) == blog.Path+"register" {
		hbx.Set("Setup", true)
	}

	file, err := ioutil.ReadFile("./blog/views/login.handlebars")
	response.appendError(err)

	hbx.Set("Blog", blog)
	response.writeResponse(hbx)

	result, err := velvet.Render(string(file), hbx)
	handleError(err)

	ctx.SetContentType("text/html")
	ctx.WriteString(result)
}

func (blog *Blog) logoutRoute(ctx *fasthttp.RequestCtx) {
	cookie := fasthttp.AcquireCookie()
	expires := time.Now()

	cookie.SetKey("JWT")
	cookie.SetValue("")
	cookie.SetExpire(expires)
	ctx.Response.Header.SetCookie(cookie)

	ctx.Redirect("/blog/", fasthttp.StatusTemporaryRedirect)
}

/* POST ROUTES */
func (blog *Blog) loginRoute(ctx *fasthttp.RequestCtx) {
	var response Response
	var user *User

	args := ctx.PostBody()
	err := json.Unmarshal(args, &user)
	response.appendError(err)

	test := blog.getUserByUsername(user.Username)

	if test.Password == user.Password {
		token, expires := blog.createJWTToken([]byte(user.Username), []byte(user.Password))
		cookie := fasthttp.AcquireCookie()
		cookie.SetKey("JWT")
		cookie.SetValue(token)
		cookie.SetExpire(expires)
		ctx.Response.Header.SetCookie(cookie)

		response.setResponse("success")

		response.writeResponse(ctx)
	} else {
		response.setResponse("Invalid password, please try again.")
		response.writeResponse(ctx)
	}
}

func (blog *Blog) registerRoute(ctx *fasthttp.RequestCtx) {
	var user *User
	var response Response
	userCount := blog.getUserCount()

	if userCount == 0 {
		args := ctx.PostBody()

		err := json.Unmarshal(args, &user)
		if err != nil {
			response.appendError(err)
		} else {
			err = blog.createUser(user.Username, user.Password)
			if err != nil {
				response.appendError(err)
			} else {
				response.setResponse("success")
			}
		}
	} else {
		response.setResponse("failed")
	}
	response.writeResponse(ctx)
}

func (blog *Blog) createPostRoute(ctx *fasthttp.RequestCtx) {
	var response Response
	var location []byte
	var post *Post

	err := blog.validateJWTTokenMiddleware(ctx)
	if err != nil {
		response.appendError(err)
	} else {
		args := ctx.PostBody()

		err = json.Unmarshal(args, &post)
		if err != nil {
			response.appendError(err)
		} else {
			ID, err := blog.createPost(post)

			if err != nil {
				response.appendError(err)
			} else {
				location = append(location, []byte(blog.Path)...)
				location = append(location, []byte("post/")...)
				strid := strconv.Itoa(int(ID))
				location = append(location, []byte(strid)...)
				response.setResponse(string(location))
			}
		}
	}
	response.writeResponse(ctx)
}

func (blog *Blog) updatePostRoute(ctx *fasthttp.RequestCtx) {
	var response Response
	var location []byte
	var post *Post

	err := blog.validateJWTTokenMiddleware(ctx)
	if err != nil {
		response.appendError(err)
	} else {
		args := ctx.PostBody()

		err = json.Unmarshal(args, &post)
		if err != nil {
			response.appendError(err)
		} else {
			err = blog.updatePost(post)

			if err != nil {
				response.appendError(err)
			} else {
				location = append(location, []byte(blog.Path)...)
				location = append(location, []byte("post/")...)
				strid := strconv.Itoa(int(post.ID))
				location = append(location, []byte(strid)...)
				response.setResponse(string(location))
			}
		}
	}
	response.writeResponse(ctx)
}

func (blog *Blog) deletePostRoute(ctx *fasthttp.RequestCtx) {
	var response Response
	var location []byte

	err := blog.validateJWTTokenMiddleware(ctx)
	if err != nil {
		response.appendError(err)
	} else {
		_ID, _ := ctx.UserValue("ID").(string)
		ID := 0
		if _ID != "" {
			ID, err = strconv.Atoi(_ID)
		}
		if err != nil {
			response.appendError(err)
		} else {
			err = blog.deletePost(ID)
			if err != nil {
				response.appendError(err)
			} else {
				location = append(location, []byte(blog.Path)...)
				location = append(location, []byte("posts/")...)
				response.Response = string(location)
			}
		}
	}
	response.writeResponse(ctx)
}

/* STATIC ROUTES */

func (blog *Blog) jsRoute(ctx *fasthttp.RequestCtx) {
	path := fmt.Sprintf("."+blog.Path+"/js/%s", ctx.UserValue("file"))
	ctx.SendFile(path)
}

func (blog *Blog) styleRoute(ctx *fasthttp.RequestCtx) {
	path := fmt.Sprintf("."+blog.Path+"/styles/%s", ctx.UserValue("file"))
	ctx.SendFile(path)
}

// Setup sets up all the routes
func (blog *Blog) Setup(router *router.Router, database string) error {
	err := blog.initialiseDatabase(database)

	/* USER ROUTES */

	/* GET ROUTES */
	router.GET(blog.Path, blog.indexRoute)
	router.GET(blog.Path+"posts", blog.indexRoute)
	router.GET(blog.Path+"posts/{pagenumber}", blog.indexRoute)
	router.GET(blog.Path+"post/{ID}", blog.postRoute)
	router.GET(blog.Path+"js/{file}", blog.jsRoute)
	router.GET(blog.Path+"styles/{file}", blog.styleRoute)
	router.GET(blog.Path+"login", blog.loginViewRoute)
	router.GET(blog.Path+"logout", blog.logoutRoute)
	router.GET(blog.Path+"register", blog.loginViewRoute)

	/* POST ROUTES */
	router.POST(blog.Path+"login", blog.loginRoute)
	router.POST(blog.Path+"register", blog.registerRoute)

	/* ADMIN ROUTES */

	/* GET ROUTES */
	router.GET(blog.Path+"editor/", blog.editorRoute)
	router.GET(blog.Path+"editor/{ID}", blog.editorRoute)

	/* POST ROUTES */
	router.POST(blog.Path+"newPost", blog.createPostRoute)
	router.POST(blog.Path+"updatePost/{ID}", blog.updatePostRoute)
	router.POST(blog.Path+"deletePost/{ID}", blog.deletePostRoute)

	return err
}
