package main

import (
	"github.com/spf13/viper"
)

type WebConfig struct {
	Address string
	Port    string
}

type AmqpConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type Config struct {
	Web  WebConfig      `mapstructure:"web"`
	Db   DatabaseConfig `mapstructure:"database"`
	Amqp AmqpConfig     `mapstructure:"amqp"`
}

var config Config

func init() {
	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := v.Unmarshal(&config); err != nil {
		panic(err)
	}
}

/*
Copyright 2020 Lorenzo Cabrini

Use of this source code is governed by an MIT-style
license that can be found in the LICENSE file or at
https://opensource.org/licenses/MIT.
*/
