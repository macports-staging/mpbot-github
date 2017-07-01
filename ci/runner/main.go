package main

import (
	"log"

	"github.com/macports/mpbot-github/ci"
)

func main() {
	session, err := ci.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	err = session.Run()
	ci.GlobalLogger.Quit()
	if err != nil {
		log.Fatal(err)
	}
}
