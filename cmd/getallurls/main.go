package main

import (
	"fmt"
	"log"

	"github.com/mgilbir/elecciones"
)

//TODO: Add flags to configure filename and refresh time
func main() {
	conf, err := elecciones.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	for n := range conf.Walk() {
		fmt.Println(n.URL())
	}
}
