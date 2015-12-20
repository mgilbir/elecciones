package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

var (
	paisURL      = "http://resultadosgenerales2015.interior.es/congreso/config/ES201512-CON-ES/pais.json"
	comunidadURL = "http://resultadosgenerales2015.interior.es/congreso/config/ES201512-CON-ES/comunidad.json"
	provinciaURL = "http://resultadosgenerales2015.interior.es/congreso/config/ES201512-CON-ES/provincia.json"
	islasURL     = "http://resultadosgenerales2015.interior.es/congreso/config/ES201512-CON-ES/islas.json"
	municipioURL = "http://resultadosgenerales2015.interior.es/congreso/config/ES201512-CON-ES/municipio.json"
	distritoURL  = "http://resultadosgenerales2015.interior.es/congreso/config/ES201512-CON-ES/distrito.json"

	urlFormat = "http://resultadosgenerales2015.interior.es/congreso/results/ES201512-CON-ES/%s/info.json"
)

type Pais []string

func (p Pais) ParentID() string {
	return p[2]
}

func (p Pais) Name() string {
	return p[1]
}

func (p Pais) ID() string {
	return p[0]
}

type Paises []Pais

type Comunidad []string

func (p Comunidad) ParentID() string {
	return p[2]
}

func (p Comunidad) Name() string {
	return p[1]
}

func (p Comunidad) ID() string {
	return p[0]
}

type Comunidades []Comunidad

type Provincia []string

func (p Provincia) ParentID() string {
	return p[2]
}

func (p Provincia) Name() string {
	return p[1]
}

func (p Provincia) ID() string {
	return p[0]
}

type Provincias []Provincia

type Isla []string

func (p Isla) ParentID() string {
	return p[2]
}

func (p Isla) Name() string {
	return p[1]
}

func (p Isla) ID() string {
	return p[0]
}

type Islas []Isla

type Municipio []string

func (p Municipio) ParentID() string {
	return p[2]
}

func (p Municipio) Name() string {
	return p[1]
}

func (p Municipio) ID() string {
	return p[0]
}

type Municipios []Municipio

type Distrito []string

func (p Distrito) ParentID() string {
	return p[2]
}

func (p Distrito) Name() string {
	return p[1]
}

func (p Distrito) ID() string {
	return p[0]
}

type Distritos []Distrito

type RawNode interface {
	Name() string
	ID() string
	ParentID() string
}

type Node struct {
	data     RawNode
	parent   *Node
	children []*Node
}

func NewNode(node RawNode, parent *Node) (*Node, error) {
	return &Node{
		data:   node,
		parent: parent,
	}, nil
}

func (n *Node) AddChild(node *Node) {
	n.children = append(n.children, node)
}

func (n Node) Parent() *Node {
	return n.parent
}

func (n Node) Children() []*Node {
	return n.children
}

func reverseStringSlice(a []string) []string {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
	return a
}

func (n Node) Path() string {
	items := []string{n.data.ID()}

	for s := n.Parent(); s != nil; s = s.Parent() {
		items = append(items, s.data.ID())
	}
	items = reverseStringSlice(items)

	return strings.Join(items, "/")
}

func (n Node) URL() string {
	return fmt.Sprintf(urlFormat, n.Path())
}

type Config struct {
	paises      map[string]*Node
	comunidades map[string]*Node
	provincias  map[string]*Node
	islas       map[string]*Node
	municipios  map[string]*Node
	distritos   map[string]*Node
}

func (c *Config) AddPais(n *Node) {
	c.paises[n.data.ID()] = n
}

func (c *Config) AddComunidad(n *Node) {
	c.comunidades[n.data.ID()] = n
	c.paises[n.Parent().data.ID()].AddChild(n)
}

func (c *Config) AddProvincia(n *Node) {
	c.provincias[n.data.ID()] = n
	c.comunidades[n.Parent().data.ID()].AddChild(n)
}

func (c *Config) AddIsla(n *Node) {
	c.islas[n.data.ID()] = n
	c.provincias[n.Parent().data.ID()].AddChild(n)
}

func (c *Config) AddMunicipio(n *Node) {
	c.municipios[n.data.ID()] = n
	parent, ok := c.provincias[n.Parent().data.ID()]
	if !ok {
		parent, ok = c.islas[n.Parent().data.ID()]
		if !ok {
			log.Printf("Cannot find parent for municipio with ID %s (%s)\n", n.data.ID(), n.data.Name())
			return
		}
	}
	parent.AddChild(n)
}

func (c *Config) AddDistrito(n *Node) {
	c.distritos[n.data.ID()] = n
	c.municipios[n.Parent().data.ID()].AddChild(n)
}

func (c Config) Walk() chan *Node {
	ch := make(chan *Node)
	go func(ch chan *Node) {
		defer close(ch)
		var wg sync.WaitGroup

		wg.Add(1)
		go func(c Config, ch chan *Node, wg *sync.WaitGroup) {
			defer wg.Done()

			for _, n := range c.paises {
				ch <- n
			}
		}(c, ch, &wg)

		wg.Add(1)
		go func(c Config, ch chan *Node, wg *sync.WaitGroup) {
			defer wg.Done()

			for _, n := range c.comunidades {
				ch <- n
			}
		}(c, ch, &wg)

		wg.Add(1)
		go func(c Config, ch chan *Node, wg *sync.WaitGroup) {
			defer wg.Done()

			for _, n := range c.provincias {
				ch <- n
			}
		}(c, ch, &wg)

		wg.Add(1)
		go func(c Config, ch chan *Node, wg *sync.WaitGroup) {
			defer wg.Done()

			for _, n := range c.islas {
				ch <- n
			}
		}(c, ch, &wg)

		wg.Add(1)
		go func(c Config, ch chan *Node, wg *sync.WaitGroup) {
			defer wg.Done()

			for _, n := range c.municipios {
				ch <- n
			}
		}(c, ch, &wg)

		wg.Add(1)
		go func(c Config, ch chan *Node, wg *sync.WaitGroup) {
			defer wg.Done()

			for _, n := range c.distritos {
				ch <- n
			}
		}(c, ch, &wg)

		wg.Wait()
	}(ch)

	return ch
}

func NewConfig() (*Config, error) {
	return &Config{
		paises:      make(map[string]*Node),
		comunidades: make(map[string]*Node),
		provincias:  make(map[string]*Node),
		islas:       make(map[string]*Node),
		municipios:  make(map[string]*Node),
		distritos:   make(map[string]*Node),
	}, nil
}

func loadPaises() (Paises, error) {
	var paises Paises
	resp, err := http.Get(paisURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(body, &paises)
	return paises, nil
}

func loadComunidades() (Comunidades, error) {
	var comunidades Comunidades
	resp, err := http.Get(comunidadURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(body, &comunidades)
	return comunidades, nil
}

func loadProvincias() (Provincias, error) {
	var provincias Provincias
	resp, err := http.Get(provinciaURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(body, &provincias)
	return provincias, nil
}

func loadIslas() (Islas, error) {
	var islas Islas
	resp, err := http.Get(islasURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(body, &islas)
	return islas, nil
}

func loadMunicipios() (Municipios, error) {
	var municipios Municipios
	resp, err := http.Get(municipioURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(body, &municipios)
	return municipios, nil
}

func loadDistritos() (Distritos, error) {
	var distritos Distritos
	resp, err := http.Get(distritoURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(body, &distritos)
	return distritos, nil
}

func loadConfig() (*Config, error) {
	conf, err := NewConfig()
	if err != nil {
		return nil, err
	}

	paises, err := loadPaises()
	if err != nil {
		return conf, err
	}

	comunidades, err := loadComunidades()
	if err != nil {
		return conf, err
	}

	provincias, err := loadProvincias()
	if err != nil {
		return conf, err

	}

	islas, err := loadIslas()
	if err != nil {
		return conf, err
	}

	municipios, err := loadMunicipios()
	if err != nil {
		return conf, err
	}

	distritos, err := loadDistritos()
	if err != nil {
		return conf, err
	}

	for _, c := range paises {
		n, err := NewNode(c, nil)
		if err != nil {
			return nil, err
		}
		conf.AddPais(n)
	}

	for _, c := range comunidades {
		pais, ok := conf.paises[c.ParentID()]
		if !ok {
			log.Printf("Uknown 'pais ID' %s for 'comunidad': %s. Skippin\n", c.ParentID(), c.Name())
			continue
		}
		n, err := NewNode(c, pais)
		if err != nil {
			return nil, err
		}
		conf.AddComunidad(n)
	}

	for _, p := range provincias {
		comunidad, ok := conf.comunidades[p.ParentID()]
		if !ok {
			log.Printf("Uknown 'comunidad ID' %s for 'provincia': %s. Skippin\n", p.ParentID(), p.Name())
			continue
		}
		n, err := NewNode(p, comunidad)
		if err != nil {
			return nil, err
		}
		conf.AddProvincia(n)
	}

	for _, i := range islas {
		provincia, ok := conf.provincias[i.ParentID()]
		if !ok {
			log.Printf("Uknown 'provincia ID' %s for 'isla': %s. Skipping\n", i.ParentID(), i.Name())
			continue
		}
		n, err := NewNode(i, provincia)
		if err != nil {
			return nil, err
		}
		conf.AddIsla(n)
	}

	for _, m := range municipios {
		parent, ok := conf.provincias[m.ParentID()]
		if !ok {
			parent, ok = conf.islas[m.ParentID()]
			if !ok {
				log.Printf("Uknown 'provincia ID' or 'isla ID' %s for 'municipio': %s\n", m.ParentID(), m.Name())
				continue
			}
		}
		n, err := NewNode(m, parent)
		if err != nil {
			return nil, err
		}
		conf.AddMunicipio(n)
	}

	for _, d := range distritos {
		parent, ok := conf.municipios[d.ParentID()]
		if !ok {

			log.Printf("Uknown 'municipio ID' %s for 'distrito': %s\n", d.ParentID(), d.Name())
			continue
		}
		n, err := NewNode(d, parent)
		if err != nil {
			return nil, err
		}
		conf.AddDistrito(n)
	}

	return conf, nil
}

func currentTime() time.Time {
	return time.Now()
}

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

func loadDataUrl(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	// var b bytes.Buffer
	// w := gzip.NewWriter(&b)
	// _, err = io.Copy(w, resp.Body)
	// if err != nil {
	// 	return nil, err
	// }

	return ioutil.ReadAll(resp.Body)
}

func retrieveData(conf *Config, db *bolt.DB) {
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

func StatsHandleFunc(w http.ResponseWriter, req *http.Request) {
	if db == nil {
		http.Error(w, "DB not available", http.StatusInternalServerError)
	}

	err := db.View(func(tx *bolt.Tx) error {
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

func ReadHandleFunc(w http.ResponseWriter, req *http.Request) {
	if db == nil {
		http.Error(w, "DB not available", http.StatusInternalServerError)
	}

	err := db.View(func(tx *bolt.Tx) error {
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

func BackupHandleFunc(w http.ResponseWriter, req *http.Request) {
	if db == nil {
		http.Error(w, "DB not available", http.StatusInternalServerError)
	}

	err := db.View(func(tx *bolt.Tx) error {
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

var (
	db *bolt.DB
)

func main() {
	var err error
	db, err = bolt.Open("congreso20D2015.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/dbbackup", BackupHandleFunc)
	http.HandleFunc("/stats", StatsHandleFunc)
	http.HandleFunc("/testread", ReadHandleFunc)

	go http.ListenAndServe(":8080", nil)

	conf, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	go retrieveData(conf, db)

	ticker := time.NewTicker(5 * time.Minute)

	for _ = range ticker.C {
		go retrieveData(conf, db)
	}
}
