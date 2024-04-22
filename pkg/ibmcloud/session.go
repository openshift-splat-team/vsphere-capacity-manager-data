package ibmcloud

import (
	"log"

	"github.com/softlayer/softlayer-go/session"
)

func ibmSession() {
	username := "foo"
	token := "foo"

	sess := session.New(username, token)

	log.Print(sess)
}
