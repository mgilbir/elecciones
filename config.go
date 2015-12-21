package elecciones

import (
	"log"
	"sync"
)

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
