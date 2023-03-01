package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleAuth             = &oauth2.Config{
		RedirectURL: "http://localhost:3000/googlecallback",
		Scopes:      []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:    google.Endpoint,
	}
	randomState = "random"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(".env file loading error - ", err)
	}
	googleAuth.ClientID = os.Getenv("GOOGLE_CLIENT_ID")
	googleAuth.ClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
}

type Config struct{}

var app *Config

func main() {

	fmt.Println(googleAuth)
	http.HandleFunc("/", app.Home)
	http.HandleFunc("/googlelogin", app.GoogleLogin)
	http.HandleFunc("/googlecallback", app.GoogleCallback)

	fmt.Println("Server starting at port 3000")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Panic(err)
	}
}
func render(w http.ResponseWriter, t string) {

	partials := []string{
		"./web/static/index.html",
	}

	var templateSlice []string
	templateSlice = append(templateSlice, fmt.Sprintf("./web/static/%s", t))

	for i := 0; i < len(partials); i++ {
		templateSlice = append(templateSlice, partials...)
	}

	tmpl, err := template.ParseFiles(templateSlice...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Config) Home(w http.ResponseWriter, r *http.Request) {
	render(w, "index.html")
}

func (app *Config) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleAuth.AuthCodeURL(randomState)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (app *Config) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != randomState {
		var html = "<html><body><h3>State is not valid</h3></body></hmtl>"
		fmt.Fprintln(w, html)
		return
	}
	fmt.Println(r.FormValue("state"), r.FormValue("code"))
	token, err := googleAuth.Exchange(context.TODO(), r.FormValue("code"))
	if err != nil {
		log.Println(err)
		var html = "<html><body><h3>Error while exchange</h3></body></hmtl>"
		fmt.Fprintln(w, html)
		return
	}
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		log.Println(err)
		var html = "<html><body><h3>Could not create get request</h3></body></hmtl>"
		fmt.Fprintln(w, html)
		return
	}

	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		var html = "<html><body><h3>Could not parse response</h3></body></hmtl>"
		fmt.Fprintln(w, html)
		return
	}
	fmt.Fprintf(w, "Respose:%s", content)
}
