package models

import (
	"time"

	"gopkg.in/mgo.v2"
)

func MongoConn(url string) {
	session, err := mgo.DialWithTimeout(url, 10*time.Second)
	if err != nil {
		panic(err)
	}
	MongoSession = mongoSession{session}
}

type mongoSession struct {
	session *mgo.Session
}

func (m *mongoSession) Get() *mgo.Session {
	return m.session.Copy()
}

var MongoSession mongoSession
