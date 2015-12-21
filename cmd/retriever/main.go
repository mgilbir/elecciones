package main

import (
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/mgilbir/elecciones"
)

//TODO: Add flags to configure filename and refresh time
func main() {
	db, err := bolt.Open("congreso20D2015.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.Handle("/stats", elecciones.NewStatsHandler(db))
	http.Handle("/dbbackup", elecciones.NewBackupHandler(db))
	http.Handle("/testread", elecciones.NewReadHandler(db))

	go http.ListenAndServe(":8080", nil)

	conf, err := elecciones.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	go elecciones.RetrieveData(conf, db)

	ticker := time.NewTicker(5 * time.Minute)

	for _ = range ticker.C {
		go elecciones.RetrieveData(conf, db)
	}
}
