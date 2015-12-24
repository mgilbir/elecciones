package elecciones

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
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

type ProgressInfo struct {
	Processed int   `json:"processed"`
	Timestamp int64 `json:"timestamp"`
	Total     int   `json:"total"`
}

type Result struct {
	AbstPercent    json.Number   `json:"abstPercent"`
	Abstention     int           `json:"abstention"`
	Blank          int           `json:"blank"`
	Census         int           `json:"census"`
	CountedCensus  int           `json:"countedCensus"`
	CountedPercent json.Number   `json:"countedPercent"`
	Null           int           `json:"null"`
	Parties        []PartyResult `json:"parties"`
	Voters         int           `json:"voters"`
}

func (r Result) ExportToCsv(partiesHeaderOrder []string, withAbsolute bool, withPercentage bool, withSeats bool) []string {
	var out []string

	out = append(out, strconv.Itoa(r.Census))
	out = append(out, strconv.Itoa(r.CountedCensus))
	out = append(out, r.CountedPercent.String())
	out = append(out, strconv.Itoa(r.Voters))
	out = append(out, strconv.Itoa(r.Abstention))
	out = append(out, strconv.Itoa(r.Blank))
	out = append(out, strconv.Itoa(r.Null))

	out = append(out, r.ExportPartiesToCsv(partiesHeaderOrder, withAbsolute, withPercentage, withSeats)...)

	return out
}

func (r Result) GetCsvHeaders(partiesHeaderOrder []string, withAbsolute bool, withPercentage bool, withSeats bool) []string {
	var out []string

	out = append(out, "census")
	out = append(out, "counted_census")
	out = append(out, "counted_percent")
	out = append(out, "voters")
	out = append(out, "abstention")
	out = append(out, "blank")
	out = append(out, "null")

	out = append(out, r.GetPartiesCsvHeaders(partiesHeaderOrder, withAbsolute, withPercentage, withSeats)...)

	return out
}

func (r Result) ExportPartiesToCsv(partiesHeaderOrder []string, withAbsolute bool, withPercentage bool, withSeats bool) []string {
	m := make(map[string]PartyResult)
	for _, p := range r.Parties {
		m[p.Acronym] = p
	}

	var absolutes []string
	var percentages []string
	var seats []string

	for _, h := range partiesHeaderOrder {
		var absoluteCount string
		var percentage string
		var seatCount string

		if v, ok := m[h]; ok {
			absoluteCount = strconv.Itoa(v.Votes.Presential)
			percentage = v.Votes.Percent.String()
			seatCount = strconv.Itoa(v.Seats)
		}

		absolutes = append(absolutes, absoluteCount)
		percentages = append(percentages, percentage)
		seats = append(seats, seatCount)
	}

	var out []string

	if withAbsolute {
		out = append(out, absolutes...)
	}

	if withPercentage {
		out = append(out, percentages...)
	}

	if withSeats {
		out = append(out, seats...)
	}

	return out
}

func (r Result) GetPartiesCsvHeaders(partiesHeaderOrder []string, withAbsolute bool, withPercentage bool, withSeats bool) []string {
	var out []string

	if withAbsolute {
		for _, v := range partiesHeaderOrder {
			out = append(out, "votes:"+v)
		}
	}

	if withPercentage {
		for _, v := range partiesHeaderOrder {
			out = append(out, "percentage:"+v)
		}
	}

	if withSeats {
		for _, v := range partiesHeaderOrder {
			out = append(out, "seats:"+v)
		}
	}

	return out
}

type PartyResult struct {
	Acronym string   `json:"acronym"`
	Code    string   `json:"code"`
	Color   string   `json:"color"`
	Id      string   `json:"id"`
	Members []string `json:"members"`
	Name    string   `json:"name"`
	Ord     int      `json:"ord"`
	Seats   int      `json:"seats"`
	Votes   Votes    `json:"votes"`
}

type Votes struct {
	Percent    json.Number `json:"percent"`
	Presential int         `json:"presential"`
}

type HistoricResult struct {
	Year   int `json:"year"`
	Result `json:",inline"`
}

type CurrentResult struct {
	Result `json:",inline"`
}

type Response struct {
	Historic []HistoricResult `json:"historic"`
	Progress ProgressInfo     `json:"progress"`
	Results  CurrentResult    `json:"results"`
}

func (r Response) ExportCurrentToCsv(partiesHeaderOrder []string, withAbsolute bool, withPercentage bool, withSeats bool) []string {
	var out []string

	ts := time.Unix(r.Progress.Timestamp/1000, 0)
	out = append(out, strconv.Itoa(int(ts.Unix())))
	out = append(out, strconv.Itoa(r.Progress.Total))
	out = append(out, strconv.Itoa(r.Progress.Processed))

	out = append(out, r.Results.ExportToCsv(partiesHeaderOrder, withAbsolute, withPercentage, withSeats)...)

	return out
}

func (r Response) GetCsvHeaders(partiesHeaderOrder []string, withAbsolute bool, withPercentage bool, withSeats bool) []string {
	var out []string

	out = append(out, "timestamp")
	out = append(out, "progress_total")
	out = append(out, "progress_processed")

	out = append(out, r.Results.GetCsvHeaders(partiesHeaderOrder, withAbsolute, withPercentage, withSeats)...)

	return out
}
