package main

func main() {
	configureLogging()
	connectDb()
	initUsers()
	go startAmqp()
	go startHttp()
	select {}
}

/*
Copyright 2020 Lorenzo Cabrini

Use of this source code is governed by an MIT-style
license that can be found in the LICENSE file or at
https://opensource.org/licenses/MIT.
*/
