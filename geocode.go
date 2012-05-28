// Package geocode is an interface to The Google Geocoding API.
//
// See http://code.google.com/apis/maps/documentation/geocoding/ for details.
package geocode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const api = "http://maps.googleapis.com/maps/api/geocode/json"

type Request struct {
	// One (and only one) of these must be set.
	Address  string
	Location *Point

	// Optional fields.
	Bounds   *Bounds // Lookup within this viewport.
	Region   string
	Language string

	Sensor bool

	values url.Values
}

func (r *Request) Values() url.Values {
	if r.values == nil {
		r.values = make(url.Values)
	}
	var v = r.values
	if r.Address != "" {
		v.Set("address", r.Address)
	} else if r.Location != nil {
		v.Set("latlng", r.Location.String())
	} else {
		panic("neither Address nor Location set")
	}
	if r.Bounds != nil {
		v.Set("bounds", r.Bounds.String())
	}
	if r.Region != "" {
		v.Set("region", r.Region)
	}
	if r.Language != "" {
		v.Set("language", r.Language)
	}
	v.Set("sensor", strconv.FormatBool(r.Sensor))
	return v
}

// Lookup makes the Request to the Google Geocoding API servers using
// the provided transport (or http.DefaultTransport if nil).
func (r *Request) Lookup(transport http.RoundTripper) (*Response, error) {
	if r == nil {
		panic("Lookup on nil *Request")
	}

	c := http.Client{Transport: transport}
	u := fmt.Sprintf("%s?%s", api, r.Values().Encode())
	getResp, err := c.Get(u)
	if err != nil {
		return nil, err
	}
	defer getResp.Body.Close()

	resp := new(Response)
	err = json.NewDecoder(getResp.Body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type Response struct {
	Status  string
	Results []*Result
}

type Result struct {
	Address      string         `json:"formatted_address"`
	AddressParts []*AddressPart `json:"address_components"`
	Geometry     *Geometry
	Types        []string
}

type AddressPart struct {
	Name      string `json:"long_name"`
	ShortName string `json:"short_name"`
	Types     []string
}

type Geometry struct {
	Bounds   Bounds
	Location Point
	Type     string
	Viewport Bounds
}

type Bounds struct {
	NorthEast, SouthWest Point
}

func (b Bounds) String() string {
	return fmt.Sprintf("%v|%v", b.NorthEast, b.SouthWest)
}

type Point struct {
	Lat, Lng float64
}

func (p Point) String() string {
	return fmt.Sprintf("%g,%g", p.Lat, p.Lng)
}
