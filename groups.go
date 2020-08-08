package main

import (
	"github.com/google/uuid"
)

type Group struct {
	Id        uuid.UUID `json:id`
	Groupname string    `json:groupname`
}
