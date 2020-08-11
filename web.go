package main

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"net/http"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/context"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
)

const (
	CookieName = "ttb-auth"
)

var (
	sessionStore *sessions.CookieStore
)

var fns = template.FuncMap{
	"first": func(x int, a interface{}) bool {
		return x == 0
	},
	"last": func(x int, a interface{}) bool {
		return x == reflect.ValueOf(a).Len()-1
	},
}

func indexGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session, err := sessionStore.Get(r, CookieName)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, ok := session.Values["user"].(User)
	if !ok {
		user = User{}
	}

	if user.Id == uuid.Nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	files := []string{
		"tpl/index.gohtml",
		"tpl/navbar.gohtml",
		"tpl/page.gohtml",
		"tpl/base.gohtml",
	}
	t, err := template.ParseFiles(files...)
	if err != nil {
		log.Error(err)
	}

	err = t.Execute(w, nil)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loginGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session, err := sessionStore.Get(r, CookieName)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{}
	data["hidenav"] = true
	if flashes := session.Flashes(); len(flashes) > 0 {
		data["error"] = flashes[0]
	}

	err = session.Save(r, w)
	if err != nil {
		log.Error(err)
	}

	files := []string{
		"tpl/login.gohtml",
		"tpl/base.gohtml",
	}
	t, err := template.ParseFiles(files...)
	if err != nil {
		log.Error(err)
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loginPost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var redirectUrl string

	session, err := sessionStore.Get(r, CookieName)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := authenticate(r.FormValue("username"), r.FormValue("password"))
	switch {
	case err == ErrAuthenticationFailed:
		session.AddFlash("Authentication failed")
		redirectUrl = "/login"
	case err == ErrUserInactive:
		session.AddFlash("User is inactive")
		redirectUrl = "/login"
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	default:
		redirectUrl = "/"
	}

	session.Values["user"] = user
	err = session.Save(r, w)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, redirectUrl, http.StatusFound)
}

func logoutGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session, err := sessionStore.Get(r, CookieName)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["user"] = nil
	err = session.Save(r, w)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func userListGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	data := map[string]interface{}{}
	users, _ := userList()
	data["users"] = users

	files := []string{
		"tpl/user-list.gohtml",
		"tpl/navbar.gohtml",
		"tpl/page.gohtml",
		"tpl/base.gohtml",
	}
	t, err := template.New("user-list.gohtml").Funcs(fns).ParseFiles(files...)
	if err != nil {
		log.Error(err)
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func userAddGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session, err := sessionStore.Get(r, CookieName)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{}
	if flashes := session.Flashes(); len(flashes) > 0 {
		data["errors"] = flashes
	}

	data["new"] = true
	if user, ok := session.Values["form"]; ok {
		data["user"] = user
		delete(session.Values, "form")
	}

	err = session.Save(r, w)
	if err != nil {
		log.Error(err)
	}

	files := []string{
		"tpl/user-form.gohtml",
		"tpl/navbar.gohtml",
		"tpl/page.gohtml",
		"tpl/base.gohtml",
	}
	t, err := template.ParseFiles(files...)
	if err != nil {
		log.Error(err)
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func userAddPost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var formErrors []string

	session, err := sessionStore.Get(r, CookieName)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	switch {
	case username == "":
		formErrors = append(formErrors, "no username specified")
	case usernameExists(username):
		formErrors = append(formErrors, "that username already exists")
	}

	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" {
		formErrors = append(formErrors, "no email specified")
	}

	phone := strings.TrimSpace(r.FormValue("phone"))

	password := r.FormValue("password")
	password2 := r.FormValue("password2")
	switch {
	case password == "":
		formErrors = append(formErrors, "empty password")
	case password != password2:
		formErrors = append(formErrors, "the passwords don't match")
	}

	user := User{
		Username: username,
		Password: password,
		Email:    email,
		Phone:    phone,
	}

	if len(formErrors) == 0 {
		user, err = storeUser(user)
		if err != nil {
			formErrors = append(formErrors, err.Error())
		}
	}

	if len(formErrors) > 0 {
		for _, e := range formErrors {
			session.AddFlash(e)
		}
		session.Values["form"] = user
		err = session.Save(r, w)
		if err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/users/add", http.StatusFound)
	} else {
		http.Redirect(w, r, "/users/list", http.StatusFound)
	}
}

func userGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "View: g%s", ps.ByName("id"))
}

func userEditGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "Edit: %s", ps.ByName("id"))
}

func userEditPost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "Edit: %s", ps.ByName("id"))
}

func userDeleteGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "Delete: %s", ps.ByName("id"))
}

func loginRequired(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		session, err := sessionStore.Get(r, CookieName)
		if err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		user, ok := session.Values["user"].(User)
		if !ok {
			user = User{}
		}

		if user.Id == uuid.Nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next(w, r, ps)
	}
}

func startHttp() {
	authKey := securecookie.GenerateRandomKey(64)
	encKey := securecookie.GenerateRandomKey(32)
	sessionStore = sessions.NewCookieStore(authKey, encKey)
	sessionStore.Options = &sessions.Options{
		MaxAge:   60 * 15,
		HttpOnly: true,
	}

	gob.Register(User{})

	router := httprouter.New()
	router.ServeFiles("/static/*filepath", http.Dir("./static"))
	router.GET("/", indexGet)
	router.GET("/login", loginGet)
	router.POST("/login", loginPost)
	router.GET("/logout", logoutGet)
	router.GET("/users/list", loginRequired(userListGet))
	router.GET("/users/add", loginRequired(userAddGet))
	router.POST("/users/add", loginRequired(userAddPost))
	router.GET("/users/view/:id", userGet)
	router.GET("/users/edit/:id", userEditGet)
	router.POST("/users/edit/:id", userEditPost)
	router.GET("/users/delete/:id", userDeleteGet)
	c := config.Web
	listen := fmt.Sprintf("%s:%s", c.Address, c.Port)
	log.Fatal(http.ListenAndServe(listen, context.ClearHandler(router)))
}

/*
Copyright 2020 Lorenzo Cabrini

Use of this source code is governed by an MIT-style
license that can be found in the LICENSE file or at
https://opensource.org/licenses/MIT.
*/
