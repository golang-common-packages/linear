package main

import (
	"log"

	"github.com/golang-common-packages/linear"
)

func main() {
	linearClient := linear.New(1024, false)
	if err := linearClient.Push("1", "a"); err != nil {
		log.Fatalln(err)
	}
}
