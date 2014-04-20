// Package geocode is an interface to mapping (google or OSM) APIs.
//  == Google: http://code.google.com/apis/maps/documentation/geocoding/
//  == OSM: http://wiki.openstreetmap.org/wiki/Nominatim
package geocode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const GOOGLE = "http://maps.googleapis.com/maps/api/geocode/json"
const OSM = "http://open.mapquestapi.com/nominatim/v1/reverse.php"

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

type Request struct {
	Provider string

	// One (and only one) of these must be set.
	Address  string
	Location *Point

	// Optional fields.
	Bounds   *Bounds // Lookup within this viewport.
	Region   string
	Language string

	Sensor bool
	Limit  int64

	values url.Values
}

func (r *Request) Values() url.Values {
	if r.values == nil {
		r.values = make(url.Values)
	}
	var v = r.values
	if r.Address != "" {
		switch r.Provider {
		case GOOGLE:
			v.Set("address", r.Address)
		case OSM:
			v.Set("q", r.Address)
		}
	} else if r.Location != nil {
		switch r.Provider {
		case GOOGLE:
			v.Set("latlng", r.Location.String())
		case OSM:
			v.Set("lat", fmt.Sprintf("%g", r.Location.Lat))
			v.Set("lon", fmt.Sprintf("%g", r.Location.Lng))
		}
	} else {
		panic("neither Address nor Location set")
	}
	if r.Bounds != nil {
		v.Set("bounds", r.Bounds.String())
	}

	if r.Provider == GOOGLE {
		if r.Region != "" {
			v.Set("region", r.Region)
		}
		if r.Language != "" {
			v.Set("language", r.Language)
		}
		v.Set("sensor", strconv.FormatBool(r.Sensor))
	} else {
		v.Set("limit", strconv.FormatInt(r.Limit, 10))
		v.Set("format", "json")
	}

	return v
}

// Lookup makes the Request to the Google Geocoding API servers using
// the provided transport (or http.DefaultTransport if nil).
func (r *Request) Lookup(transport http.RoundTripper) (*Response, error) {
	if r == nil {
		panic("Lookup on nil *Request")
	}

	c := http.Client{Transport: transport}
	u := fmt.Sprintf("%s?%s", r.Provider, r.Values().Encode())
	getResp, err := c.Get(u)
	if err != nil {
		return nil, err
	}
	defer getResp.Body.Close()

	resp := new(Response)
	resp.QueryString = u

	if getResp.StatusCode == 200 { // OK
		decoder := json.NewDecoder(getResp.Body)
		switch r.Provider {
		case GOOGLE:
			gResp := &GoogleResponse{Response: resp}
			// reverse geocoding
			err = decoder.Decode(gResp)
			resp.Count = len(gResp.Results)
			if resp.Count >= 1 {
				resp.Found = gResp.Results[0].Address
			}
		case OSM:
			oResp := &OSMResponse{Response: resp}
			// reverse geocoding
			err = decoder.Decode(oResp)
			if oResp.Address != "" {
				resp.Count = 1
				resp.Found = oResp.AddressParts.Name
				fmt.Println(u, oResp.Address)
			} else {
				resp.Count = 0
			}
			// geocoding
			// bodyBytes, _ := ioutil.ReadAll(getResp.Body)
			// bodyString := string(bodyBytes)
			// err = json.NewDecoder(bytes.NewBufferString(strings.TrimRight(bodyString, "]")[1:])).Decode(resp)
		}
	}

	if err != nil {
		return nil, err
	}

	resp.Status = "OK"
	return resp, nil
}

type Response struct {
	Status      string
	QueryString string
	Found       string
	Count       int
}

type GoogleResponse struct {
	*Response
	Results []*GoogleResult
}

type GoogleResult struct {
	/*
		{
		   "results" : [
		      {
		         "address_components" : [
		            {
		               "long_name" : "1600",
		               "short_name" : "1600",
		               "types" : [ "street_number" ]
		            },
		            {
		               "long_name" : "Amphitheatre Pkwy",
		               "short_name" : "Amphitheatre Pkwy",
		               "types" : [ "route" ]
		            },
		            {
		               "long_name" : "Mountain View",
		               "short_name" : "Mountain View",
		               "types" : [ "locality", "political" ]
		            },
		            {
		               "long_name" : "Santa Clara",
		               "short_name" : "Santa Clara",
		               "types" : [ "administrative_area_level_2", "political" ]
		            },
		            {
		               "long_name" : "California",
		               "short_name" : "CA",
		               "types" : [ "administrative_area_level_1", "political" ]
		            },
		            {
		               "long_name" : "United States",
		               "short_name" : "US",
		               "types" : [ "country", "political" ]
		            },
		            {
		               "long_name" : "94043",
		               "short_name" : "94043",
		               "types" : [ "postal_code" ]
		            }
		         ],
		         "formatted_address" : "1600 Amphitheatre Pkwy, Mountain View, CA 94043, USA",
		         "geometry" : {
		            "location" : {
		               "lat" : 37.42291810,
		               "lng" : -122.08542120
		            },
		            "location_type" : "ROOFTOP",
		            "viewport" : {
		               "northeast" : {
		                  "lat" : 37.42426708029149,
		                  "lng" : -122.0840722197085
		               },
		               "southwest" : {
		                  "lat" : 37.42156911970850,
		                  "lng" : -122.0867701802915
		               }
		            }
		         },
		         "types" : [ "street_address" ]
		      }
		   ],
		   "status" : "OK"
		}
	*/
	Address      string               `json:"formatted_address"`
	AddressParts []*GoogleAddressPart `json:"address_components"`
	Geometry     *Geometry
	Types        []string
}

type GoogleAddressPart struct {
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

type OSMResponse struct {
	*Response
	// OSM stuff
	/*
		{"place_id":"62762024",
		"licence":"Data \u00a9 OpenStreetMap contributors, ODbL 1.0. http:\/\/www.openstreetmap.org\/copyright",
		"osm_type":"way",
		"osm_id":"90394420",
		"lat":"52.548781",
		"lon":"-1.81626870827795",
		"display_name":"137, Pilkington Avenue, Castle Vale, Birmingham, West Midlands, England, B72 1LH, United Kingdom",
		"address":{
			"house_number":"137",
			"road":"Pilkington Avenue",
			"suburb":"Castle Vale",
			"city":"Birmingham",
			"county":"West Midlands",
			"state_district":"West Midlands",
			"state":"England",
			"postcode":"B72 1LH",
			"country":"United Kingdom",
			"country_code":"gb"
		}}
	*/
	Address      string          `json:"display_name"`
	AddressParts *OSMAddressPart `json:"address"`
	Lat          string          `json:"lat"`
	Lng          string          `json:"lon"`
}

type OSMAddressPart struct {
	HouseNumber string `json:"house_number"`
	Name        string `json:"road"`
	City        string `json:"city"`
	State       string `json:"state"`
}
