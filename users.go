package main

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id        uuid.UUID `json:id`
	Username  string    `json:username`
	Password  string    `json:password`
	Email     string    `json:email`
	Phone     string    `json:phone`
	Token     uuid.UUID `json:token`
	Active    bool      `json:active`
	CreatedAt time.Time `json:created_at`
}

type UserList struct {
	Id       uuid.UUID `json:id`
	Username string    `json:username`
}

var (
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrUserInactive         = errors.New("user is inactive")
)

func storeUser(user User) (User, error) {
	if user.Id == uuid.Nil {
		return addUser(user)
	} else {
		return updateUser(user)
	}
}

func usernameExists(username string) bool {
	var count int

	q := "SELECT COUNT(*) " +
		"FROM users " +
		"WHERE username = $1"

	rows, err := db.Query(q, username)
	if err != nil {
		log.Error(err)
		// TODO: what should be done here? Do we propagate the error,
		// return false or something else?
	}
	defer rows.Close()

	rows.Next()
	rows.Scan(&count)
	return count == 1
}

func initUsers() {
	if !usernameExists("sa") {
		log.Info("creating sysadmin user")
		user := User{
			Username: "sa",
			Password: "s3kr3t",
		}
		storeUser(user)
	}
}

func authenticate(username string, password string) (User, error) {
	q := "SELECT * FROM users " +
		"WHERE username = $1 AND password = crypt($2, password)"

	user := User{}
	err := db.QueryRow(q, username, password).Scan(
		&user.Id,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.Phone,
		&user.Token,
		&user.Active,
		&user.CreatedAt)
	switch {
	case err == sql.ErrNoRows:
		return User{}, ErrAuthenticationFailed
	case err != nil:
		log.Error(err)
		return User{}, err
	}

	if user.Active == false {
		return User{}, ErrUserInactive
	}

	return user, nil
}

func addUser(user User) (User, error) {
	user.Id = uuid.New()
	user.Token = uuid.New()

	q := "INSERT INTO users" +
		"(id, username, password, token, email, phone) " +
		"VALUES($1, $2, crypt($3, gen_salt('bf')), $4, $5, $6)"

	_, err := db.Exec(
		q,
		user.Id,
		user.Username,
		user.Password,
		user.Token,
		user.Email,
		user.Phone)
	if err != nil {
		log.Error(err)
	}

	return user, err
}

func updateUser(user User) (User, error) {
	q := "UPDATE users " +
		"SET username = $1, password = crypt($2, gen_salt('bf')), " +
		"email = $3, phone = $4, token = $5, active = $6 " +
		"WHERE id = $7"
	_, err := db.Exec(
		q,
		user.Username,
		user.Password,
		user.Email,
		user.Phone,
		user.Token,
		user.Active,
		user.Id)
	if err != nil {
		log.Error(err)
	}

	return user, err
}

func userList() ([]UserList, error) {
	var users []UserList

	q := "SELECT * FROM userlist"
	rows, err := db.Query(q)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := UserList{}
		err := rows.Scan(&item.Id, &item.Username)
		if err != nil {
			log.Error(err)
			return nil, err
		} else {
			users = append(users, item)
		}
	}

	err = rows.Err()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return users, nil
}

/*
Copyright 2020 Lorenzo Cabrini

Use of this source code is governed by an MIT-style
license that can be found in the LICENSE file or at
https://opensource.org/licenses/MIT.
*/
