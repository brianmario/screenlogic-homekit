package main

import (
	"log"

	"github.com/brianmario/screenlogic-homekit/screenlogic"
)

func main() {
	gateway, err := screenlogic.DiscoverGateway()
	if err != nil {
		log.Fatal(err)
	}

	err = gateway.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer gateway.Close()

	clientName := "screenlogic-homekit"

	err = gateway.Login(clientName)
	if err != nil {
		log.Fatal(err)
	}

	_, err = gateway.Version()
	if err != nil {
		log.Fatal(err)
	}

	_, err = gateway.ControllerConfig()
	if err != nil {
		log.Fatal(err)
	}

	_, err = gateway.PoolStatus()
	if err != nil {
		log.Fatal(err)
	}
}
