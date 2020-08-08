package main

import (
	"github.com/sirupsen/logrus"
	"os"
)

var log = logrus.New()

func configureLogging() {
	// TODO: logging should be configurable
	log.Out = os.Stderr
}

/*
Copyright 2020 Lorenzo Cabrini

Use of this source code is governed by an MIT-style
license that can be found in the LICENSE file or at
https://opensource.org/licenses/MIT.
*/
