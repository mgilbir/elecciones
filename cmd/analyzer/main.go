package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/mgilbir/elecciones"
)

var (
	filename = flag.String("db", "congreso20D2015.db", "db filename")
	port     = flag.String("http", ":8081", "the port where the server is listening")
)

func main() {
	flag.Parse()

	db, err := bolt.Open(*filename, 0666, &bolt.Options{ReadOnly: true})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.Handle("/stats", elecciones.NewStatsHandler(db))

	log.Fatal(http.ListenAndServe(*port, nil))
}
