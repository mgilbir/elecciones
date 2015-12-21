package elecciones

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

func storeEntry(db *bolt.DB, n *Node, data []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(n.Path()))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		now := currentTime()
		key, err := now.MarshalBinary()
		if err != nil {
			return err
		}
		b.Put(key, data)
		return nil
	})
}

func RetrieveData(conf *Config, db *bolt.DB) {
	log.Println("Data load initiated")
	var wg sync.WaitGroup
	for n := range conf.Walk() {
		wg.Add(1)
		go func(n *Node, wg *sync.WaitGroup) {
			defer wg.Done()
			seconds := time.Duration(rand.Intn(200))
			time.Sleep(seconds * time.Second)
			data, err := loadDataUrl(n.URL())
			if err != nil {
				log.Println("Error retrieving URL %s: %v", n.URL(), err)
			}
			storeEntry(db, n, data)
		}(n, &wg)
	}
	wg.Wait()
	log.Println("Data load completed")
}
