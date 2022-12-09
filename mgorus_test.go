package mgorus

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestMgorus(t *testing.T) {
	log := logrus.New()
	hooker, err := NewHooker("mongodb://localhost:27017", "testing", "log")
	if err == nil {
		log.Hooks.Add(hooker)
	} else {
		t.Fatal(err)
	}

	log.WithFields(logrus.Fields{
		"name": "zhangsan",
		"age":  28,
	}).Error("Hello world!")
}
