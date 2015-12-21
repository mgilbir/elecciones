package elecciones

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

type StatsHandler struct {
	db *bolt.DB
}

func NewStatsHandler(db *bolt.DB) StatsHandler {
	return StatsHandler{db: db}
}

func (s StatsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if s.db == nil {
		http.Error(w, "DB not available", http.StatusInternalServerError)
	}

	err := s.db.View(func(tx *bolt.Tx) error {
		c := tx.Cursor()

		bucketCount := 0
		countPerBucket := make(map[string]int)

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			bucketCount++

			b := tx.Bucket(k)
			countPerBucket[string(k)] = b.Stats().KeyN
		}

		w.Write([]byte(fmt.Sprintf("BucketCount: %d\n", bucketCount)))
		for k, v := range countPerBucket {
			w.Write([]byte(fmt.Sprintf("Bucket %s -> %d\n", k, v)))
		}

		return nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type ReadHandler struct {
	db *bolt.DB
}

func NewReadHandler(db *bolt.DB) ReadHandler {
	return ReadHandler{db: db}
}

func (r ReadHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.db == nil {
		http.Error(w, "DB not available", http.StatusInternalServerError)
	}

	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("ES/CA02/50/50297/5029710"))
		k, v := b.Cursor().Last()

		ts := time.Time{}
		ts.UnmarshalBinary(k)

		// fmt.Println(v)
		//
		// bf := bytes.NewReader(v)
		// r, err := gzip.NewReader(bf)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return nil
		// }

		w.Write(v)
		return nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type BackupHandler struct {
	db *bolt.DB
}

func NewBackupHandler(db *bolt.DB) BackupHandler {
	return BackupHandler{db: db}
}

func (b BackupHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if b.db == nil {
		http.Error(w, "DB not available", http.StatusInternalServerError)
	}

	err := b.db.View(func(tx *bolt.Tx) error {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="my.db"`)
		w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
		_, err := tx.WriteTo(w)
		return err
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
