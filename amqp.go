package main

import (
	"encoding/json"
	"fmt"

	"github.com/streadway/amqp"
)

type AuthRequest struct {
	Username string `json:username`
	Password string `json:password`
}

type AuthResponse struct {
	Status int `json:status`
}

type UserListRequest struct {
	AppKey string `json:appkey`
}

type UserListResponse struct {
	Status int        `json:status`
	Users  []UserList `json:userlist`
}

func startAmqp() {
	var err error

	c := config.Amqp
	connStr := fmt.Sprintf("amqp://%s:%s@%s:%s",
		c.User,
		c.Password,
		c.Host,
		c.Port)

	conn, err := amqp.Dial(connStr)
	if err != nil {
		// TODO: decide what level to log
		log.Fatal(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		// TODO: decide what level to log
		log.Fatal(err)
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		// TODO: decide what level to log
		log.Fatal(err)
	}

	go consumeAuthenticate(ch)
	go consumeUserlist(ch)
	go consumeAddUser(ch)
}

func consumeAuthenticate(ch *amqp.Channel) {
	q, err := declareQueue("authenticate", ch)
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		log.Error(err)
	}

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			req := AuthRequest{}
			if err := json.Unmarshal(d.Body, &req); err != nil {
				log.Error(err)
			}

			resp := AuthResponse{}
			user, err := authenticate(req.Username, req.Password)
			switch {
			case err == ErrAuthenticationFailed:
				resp.Status = 1
			case err == ErrUserInactive:
				resp.Status = 2
			case err == nil:
				resp.Status = 0
				log.Info(user)
			default:
				resp.Status = 255
			}

			body, err := json.Marshal(resp)
			if err != nil {
				log.Error(err)
			}

			err = ch.Publish(
				"",
				d.ReplyTo,
				false,
				false,
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          body,
				})
			if err != nil {
				log.Error(err)
			}

			d.Ack(false)

		}
	}()
	<-forever
}

func consumeUserlist(ch *amqp.Channel) {
	q, err := declareQueue("listusers", ch)
	if err != nil {
		log.Error(err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		log.Error(err)
	}

	forever := make(chan bool)
	go func() {
		req := UserListRequest{}
		for d := range msgs {
			if err := json.Unmarshal(d.Body, &req); err != nil {
				log.Error(err)
			}

			resp := UserListResponse{}

			// TODO: once I add application keys, check to ensure that
			// the application key is valid.

			userList, err := userList()
			switch {
			case err != nil:
				resp.Status = 1
				resp.Users = nil
			default:
				resp.Status = 0
				resp.Users = userList
			}

			body, err := json.Marshal(resp)
			if err != nil {
				log.Error(err)
			}

			err = ch.Publish(
				"",
				d.ReplyTo,
				false,
				false,
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          body,
				})
			if err != nil {
				log.Error(err)
			}

			d.Ack(false)
		}
	}()
	<-forever

}

func consumeAddUser(ch *amqp.Channel) {

}

func declareQueue(name string, ch *amqp.Channel) (amqp.Queue, error) {
	name = "ttb.auth." + name
	q, err := ch.QueueDeclare(name, false, false, false, false, nil)
	if err != nil {
		log.Error(err)
	}
	return q, err
}

/*
Copyright 2020 Lorenzo Cabrini

Use of this source code is governed by an MIT-style
license that can be found in the LICENSE file or at
https://opensource.org/licenses/MIT.
*/
