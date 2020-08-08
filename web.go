package main

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"net/http"

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
