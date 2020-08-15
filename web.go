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
	session := getSession(w, r)
	user, ok := session.Values["user"].(User)
	if !ok {
		user = User{}
	}

	if user.Id == uuid.Nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	doTemplate("index.gohtml", nil, w)
}

func loginGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := getSession(w, r)

	data := map[string]interface{}{}
	data["hidenav"] = true
	if flashes := session.Flashes(); len(flashes) > 0 {
		data["error"] = flashes[0]
	}

	saveSession(session, w, r)

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

	session := getSession(w, r)

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
	saveSession(session, w, r)

	http.Redirect(w, r, redirectUrl, http.StatusFound)
}

func logoutGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := getSession(w, r)
	session.Values["user"] = nil
	saveSession(session, w, r)

	http.Redirect(w, r, "/", http.StatusFound)
}

func userListGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	data := map[string]interface{}{}
	users, _ := userList()
	data["users"] = users

	doTemplate("user-list.gohtml", data, w)
}

func userAddGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := getSession(w, r)

	data := map[string]interface{}{}
	if flashes := session.Flashes(); len(flashes) > 0 {
		data["errors"] = flashes
	}

	data["new"] = true
	if user, ok := session.Values["form"]; ok {
		data["user"] = user
		delete(session.Values, "form")
	}

	saveSession(session, w, r)

	doTemplate("user-form.gohtml", data, w)
}

func userAddPost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error

	session := getSession(w, r)
	formErrors := validateUserForm(r)

	user := User{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
		Email:    r.FormValue("email"),
		Phone:    r.FormValue("phone"),
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
		saveSession(session, w, r)

		http.Redirect(w, r, "/users/add", http.StatusFound)
	} else {
		http.Redirect(w, r, "/users/list", http.StatusFound)
	}
}

func userViewGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uuid, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	user := getUser(uuid)
	data := map[string]interface{}{}
	data["user"] = user
	doTemplate("user-view.gohtml", data, w)
}

func userEditGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := getSession(w, r)

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	data := map[string]interface{}{}
	if flashes := session.Flashes(); len(flashes) > 0 {
		data["errors"] = flashes
	}
	data["user"] = getUser(id)

	saveSession(session, w, r)

	doTemplate("user-form.gohtml", data, w)
}

func userEditPost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error

	session := getSession(w, r)
	formErrors := validateUserForm(r)

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	user := getUser(id)
	user.Id = id
	user.Username = r.FormValue("username")
	user.Password = r.FormValue("password")
	user.Email = r.FormValue("email")
	user.Phone = r.FormValue("phone")

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
		saveSession(session, w, r)

		http.Redirect(w, r, r.RequestURI, http.StatusFound)
	} else {
		http.Redirect(w, r, "/users/list", http.StatusFound)
	}
}

func userDeleteGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "Delete: %s", ps.ByName("id"))
}

func validateUserForm(r *http.Request) []string {
	var formErrors []string

	id, err := uuid.Parse(r.FormValue("id"))
	if err != nil {
		id = uuid.Nil
	}

	username := strings.TrimSpace(r.FormValue("username"))
	switch {
	case username == "":
		formErrors = append(formErrors, "no username specified")
	case usernameExists(username, id):
		formErrors = append(formErrors, "that username already exists")
	}

	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" {
		formErrors = append(formErrors, "no email specified")
	}

	password := r.FormValue("password")
	password2 := r.FormValue("password2")
	switch {
	case password == "":
		formErrors = append(formErrors, "empty password")
	case password != password2:
		formErrors = append(formErrors, "the passwords don't match")
	}

	return formErrors
}

func getSession(w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, err := sessionStore.Get(r, CookieName)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	return session
}

func saveSession(s *sessions.Session, w http.ResponseWriter, r *http.Request) {
	err := s.Save(r, w)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func doTemplate(tn string, data map[string]interface{}, w http.ResponseWriter) {
	files := []string{
		"tpl/" + tn,
		"tpl/navbar.gohtml",
		"tpl/page.gohtml",
		"tpl/base.gohtml",
	}
	t, err := template.New(tn).Funcs(fns).ParseFiles(files...)
	if err != nil {
		log.Error(err)
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
	router.GET("/users/view/:id", loginRequired(userViewGet))
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
