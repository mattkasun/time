package build

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var currentCookie string
var pushupGreeting = "Welcome to Pushup!"
var errorMessage string
var Version = "v0.1.0"

func auth(r *http.Request) error {
	cookie, err := r.Cookie("timetrace")
	if err != nil {
		return errors.New("not logged in")
	}
	if valid := cookie.Valid(); valid != nil {
		return valid
	}
	if cookie.Value != currentCookie {
		return errors.New("invalid cookie")
	}
	return nil
}

func login(r *http.Request) (*http.Cookie, error) {
	defer r.Body.Close()
	user := struct {
		Name string
		Pass string
	}{}
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	user.Name = r.PostFormValue("user")
	user.Pass = r.PostFormValue("pass")
	if strings.ToLower(user.Name) != "hello" || strings.ToLower(user.Pass) != "world" {
		fmt.Printf("login received user:%s pass:%s \n", user.Name, user.Pass)
		return nil, errors.New("invalid username or password")
	}
	cookie := http.Cookie{
		Name:  "timetrace",
		Value: "big test",
	}
	currentCookie = cookie.Value
	return &cookie, nil
}

func setMessage(m string) {
	errorMessage = m
}

func getError() string {
	return errorMessage
}
