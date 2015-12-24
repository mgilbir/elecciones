package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/mgilbir/elecciones"
)

var (
	filename = flag.String("db", "congreso20D2015.db", "db filename")
	port     = flag.String("http", ":8081", "the port where the server is listening")
	output   = flag.String("out", "congreso20D2015.csv", "output filename")
)

func main() {
	flag.Parse()

	db, err := bolt.Open(*filename, 0666, &bolt.Options{ReadOnly: true})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	partiesAcronyms := make(map[string]struct{})

	db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			_, v := b.Cursor().First()

			var resp elecciones.Response
			err := json.Unmarshal(v, &resp)
			if err != nil {
				return fmt.Errorf("ERROR parsing bucket %s", name)
			}
			for _, p := range resp.Results.Parties {
				partiesAcronyms[p.Acronym] = struct{}{}
			}

			return nil
		})
	})

	var headerPartyOrder []string
	for k, _ := range partiesAcronyms {
		headerPartyOrder = append(headerPartyOrder, k)
	}

	f, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.Comma = ';'
	defer w.Flush()

	headers := []string{"Pais", "Comunidad", "Provincia", "Isla", "Municipio", "Distrito"}

	headers = append(headers, elecciones.Response{}.GetCsvHeaders(headerPartyOrder, true, true, true)...)

	err = w.Write(headers)
	if err != nil {
		log.Fatal(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			return b.ForEach(func(k []byte, v []byte) error {
				var resp elecciones.Response
				err := json.Unmarshal(v, &resp)
				if err != nil {
					log.Printf("ERROR parsing bucket %s. %v\n", name, err)
				} else {

					out := getLocationArrayForBucketName(string(name))
					out = append(out, resp.ExportCurrentToCsv(headerPartyOrder, true, true, true)...)

					err = w.Write(out)
					if err != nil {
						log.Fatal(err)
					}
				}

				return nil
			})
		})
	})
	log.Println(err)
}

func getLocationArrayForBucketName(name string) []string {
	out := make([]string, 6)

	b := strings.Split(name, "/")

	out[0] = b[0]

	if len(b) > 1 {
		out[1] = b[1]
	}

	if len(b) > 2 {
		out[2] = b[2]
	}

	var isIsland bool
	if len(b) > 3 {
		if b[2] == "07" || b[2] == "35" || b[2] == "38" {
			isIsland = true
		}

		if isIsland {
			out[3] = b[3]
		} else {
			out[4] = b[3]
		}
	}

	if len(b) > 4 {
		if isIsland {
			out[4] = b[4]
		} else {
			out[5] = b[4]
		}
	}

	if len(b) > 5 {
		if isIsland {
			out[5] = b[5]
		} else {
			out[6] = b[5]
		}
	}

	return out
}
