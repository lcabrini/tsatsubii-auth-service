package main

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var db *sql.DB

func connectDb() {
	var err error

	pgInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Db.Host,
		config.Db.Port,
		config.Db.User,
		config.Db.Password,
		config.Db.Database)
	if db, err = sql.Open("postgres", pgInfo); err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: there should be a command-line option and/or configuration
	// setting to enable/disable migrations. Perhaps this should depend
	// on if the service is running in a devel or live environment.
	err = m.Up()
	switch {
	case err == migrate.ErrNoChange:
		log.Info("no migrations to run")
	case err != nil:
		log.Fatal(err)
	default:
		log.Info("database migrated")
	}
}

/*
Copyright 2020 Lorenzo Cabrini

Use of this source code is governed by an MIT-style
license that can be found in the LICENSE file or at
https://opensource.org/licenses/MIT.
*/
