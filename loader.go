package elecciones

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
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

func LoadConfig() (*Config, error) {
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
