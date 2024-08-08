package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Middleware func(http.Handler) http.Handler

type Product struct {
	Title string `json:"title"`
	Price int    `json:"price"`
}

const form = `<html>
    <head>
    <title></title>
    </head>
    <body>
        <form action="/login" method="post">
            <label>Логин</label><input type="text" name="login">
            <label>Пароль<input type="password" name="password">
            <input type="submit" value="Login">
        </form>
    </body>
</html>`

func Pipeline(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

func middleware(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		nextHandler.ServeHTTP(w, r)
	})
}

func Auth(login, password string) bool {
	return login == `guest` && password == `guest`
}

func rootHandle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Buy!"))
}

func JSONHandler(w http.ResponseWriter, r *http.Request) {
	product := Product{"Sausage", 100}

	resp, err := json.Marshal(product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func mainPage(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(`<!DOCTYPE html><html><h1>Meow!</h1></html>`))
}

func apiPageForm(res http.ResponseWriter, req *http.Request) {
	body := fmt.Sprintf("Method: %s\r\n", req.Method)
	body += "Header ===============\r\n"
	for k, v := range req.Header {
		body += fmt.Sprintf("%s: %v\r\n", k, v)
	}
	body += "Query parameters ===============\r\n"
	if err := req.ParseForm(); err != nil {
		res.Write([]byte(err.Error()))
		return
	}
	for k, v := range req.Form {
		body += fmt.Sprintf("%s: %v\r\n", k, v)
	}
	res.Write([]byte(body))
}

func apiPage(res http.ResponseWriter, req *http.Request) {
	body := fmt.Sprintf("Method: %s\r\n", req.Method)
	body += "Header ===============\r\n"
	for k, v := range req.Header {
		body += fmt.Sprintf("%s: %v\r\n", k, v)
	}
	body += "Query parameters ===============\r\n"
	for k, v := range req.URL.Query() {
		body += fmt.Sprintf("%s: %v\r\n", k, v)
	}
	res.Write([]byte(body))
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		login := r.FormValue("login")
		password := r.FormValue("password")
		if Auth(login, password) {
			io.WriteString(w, "Welcome to the meow!")
		} else {
			http.Error(w, "Wrong login or password!", http.StatusUnauthorized)
		}
		return
	} else if r.Method == http.MethodGet {
		io.WriteString(w, form)
		return
	}
	w.Write([]byte(strconv.Itoa(http.StatusMethodNotAllowed)))
}

func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://duckduckgo.com/", http.StatusMovedPermanently)
}

// type HandlerWeb struct{}

/*func (h HandlerWeb) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	data := []byte("Meow!")
	res.Write(data)
}*/

func main() {
	//var h HandlerWeb

	mux := http.NewServeMux()

	mux.HandleFunc(`/main`, mainPage)
	mux.HandleFunc(`/api`, apiPage)
	mux.HandleFunc(`/api/form`, apiPageForm)
	mux.HandleFunc(`/api/json`, JSONHandler)
	mux.HandleFunc(`/login`, login)
	mux.Handle("/", middleware(http.HandlerFunc(rootHandle)))
	mux.Handle("/pipeline", Pipeline(http.HandlerFunc(rootHandle), middleware))
	mux.HandleFunc("/search/", redirect)
	//http.Handle("/", middleware(http.HandlerFunc(rootHandle)))

	// err := http.ListenAndServe(`:8080`, http.HandlerFunc(mainPage))

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
