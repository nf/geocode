// Package geocode is an interface to mapping APIs. This includes geocoding as well as routing.
//  == Google: http://code.google.com/apis/maps/documentation/geocoding/
//  == OSM: http://wiki.openstreetmap.org/wiki/Nominatim
//  == YOURS: http://wiki.openstreetmap.org/wiki/YOURS
package geocode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type (
	RequestType         int
	ProviderApiLocation string
)

const (
	/* Request Type */
	GEOCODE RequestType = 1
	ROUTE   RequestType = 2

	/* Geocoding URLs */
	GOOGLE = "http://maps.googleapis.com/maps/api/geocode/json"
	OSM    = "http://open.mapquestapi.com/nominatim/v1/reverse.php"

	/* Routing URLs */
	YOURS        = "http://www.yournavigation.org/api/1.0/gosmore.php"
	YOURS_HEADER = "github.com/talmai/geocode" // change this to your website!
)

type Point struct {
	Lat, Lng float64
}

func (p Point) String() string {
	return fmt.Sprintf("%g,%g", p.Lat, p.Lng)
}

type Bounds struct {
	NorthEast, SouthWest Point
}

func (b Bounds) String() string {
	return fmt.Sprintf("%v|%v", b.NorthEast, b.SouthWest)
}

type Request struct {
	HTTPClient *http.Client
	Provider   ProviderApiLocation
	Type       RequestType

	// For geocoding, one (and only one) of these must be set.
	Address  string
	Location *Point

	// Api key
	ApiKey string

	// For routing, bounds must be set
	Bounds *Bounds // used by GOOGLE and YOURS

	Limit    int64  // used by OSM
	Region   string // used by GOOGLE
	Language string // used by GOOGLE
	Sensor   bool   // used by GOOGLE

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
		if r.Type == GEOCODE {
			panic("neither Address nor Location set")
		}
	}

	if r.Bounds != nil {
		switch r.Provider {
		case GOOGLE:
			v.Set("bounds", r.Bounds.String())
		case YOURS:
			v.Set("flat", fmt.Sprintf("%g", r.Bounds.NorthEast.Lat))
			v.Set("flon", fmt.Sprintf("%g", r.Bounds.NorthEast.Lng))
			v.Set("tlat", fmt.Sprintf("%g", r.Bounds.SouthWest.Lat))
			v.Set("tlon", fmt.Sprintf("%g", r.Bounds.SouthWest.Lng))
		}
	} else {
		if r.Type == ROUTE {
			panic("Start/End Bounds must be set for routing")
		}
	}

	switch r.Provider {
	case GOOGLE:
		if r.Region != "" {
			v.Set("region", r.Region)
		}
		if r.Language != "" {
			v.Set("language", r.Language)
		}
		if r.ApiKey != "" {
			v.Set("key", r.ApiKey)
		}
		v.Set("sensor", strconv.FormatBool(r.Sensor))
	case OSM:
		v.Set("limit", strconv.FormatInt(r.Limit, 10))
		v.Set("format", "json")
	case YOURS:
		v.Set("v", "motorcar")   // type of transport, possible options are: motorcar, bicycle or foot.
		v.Set("fast", "1")       // selects the fastest route, 0 the shortest route.
		v.Set("layer", "mapnik") // determines which Gosmore instance is used to calculate the route
		//	Provide mapnik for normal routing using car, bicycle or foot
		//	Provide cn for using bicycle routing using cycle route networks only.
		v.Set("format", "geojson") // This can either be kml or geojson.
		v.Set("geometry", "1")     // enables/disables adding the route geometry in the output.
		v.Set("distance", "v")     // specifies which algorithm is used to calculate the route distance
		//	Options are v for Vicenty, gc for simplified Great Circle, h for Haversine Law, cs for Cosine Law.
		v.Set("instructions", "1") // enbles/disables adding driving instructions in the output.
		v.Set("lang", "en_US")     // specifies the language code in which the routing directions are returned.
	}

	return v
}

func (r *Request) Lookup(transport http.RoundTripper) (*Response, error) {
	r.Type = GEOCODE
	return r.SendAPIRequest(transport)
}

func (r *Request) Route(transport http.RoundTripper) (*Response, error) {
	r.Type = ROUTE
	return r.SendAPIRequest(transport)
}

// SendAPIRequest makes the Request to the provider using
// the provided transport (or http.DefaultTransport if nil).
func (r *Request) SendAPIRequest(transport http.RoundTripper) (*Response, error) {
	if r == nil {
		panic("Lookup on nil *Request")
	}

	c := r.HTTPClient
	if c == nil {
		c = &http.Client{Transport: transport}
	}
	u := fmt.Sprintf("%s?%s", r.Provider, r.Values().Encode())

	req, err := http.NewRequest("GET", u, nil)
	if r.Provider == YOURS {
		req.Header.Add("X-Yours-client", YOURS_HEADER)
	}
	getResp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer getResp.Body.Close()

	resp := new(Response)
	resp.QueryString = u

	if getResp.StatusCode == 200 { // OK
		err = json.NewDecoder(getResp.Body).Decode(resp)
		switch r.Provider {
		case GOOGLE:
			// reverse geocoding
			resp.Count = len(resp.GoogleResponse.Results)
			if resp.Count >= 1 {
				resp.Found = resp.GoogleResponse.Results[0].Address
			}
		case OSM:
			// reverse geocoding
			if resp.OSMResponse.Address != "" {
				resp.Count = 1
				resp.Found = resp.OSMResponse.AddressParts.Name
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

	return resp, nil
}

type Response struct {
	Status      string
	QueryString string
	Found       string
	Count       int
	*GoogleResponse
	*OSMResponse
	*YOURSResponse
}

type GoogleResponse struct {
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

// currently doesn't
type YOURSResponse struct {
	/*
		{
		  "type": "LineString",
		  "crs": {
		    "type": "name",
		    "properties": {
		      "name": "urn:ogc:def:crs:OGC:1.3:CRS84"
		    }
		  },
		  "coordinates":
		  [
			[-118.604871, 34.172300]
			,[-118.604872, 34.172078]
			,[-118.604870, 34.171966]
			,[-118.500806, 34.235753]
			,[-118.500814, 34.236146]
		  ],
		  "properties": {
		    "distance": "17.970238",
		    "description": "Go straight ahead.<br>Follow the road for...",
		    "traveltime": "1018"
		    }
		}
	*/
	Coordinates [][]float64      `json:"coordinates"`
	Properties  *YOURSProperties `json:"properties"`
}

type YOURSProperties struct {
	Distance     string `json:"distance"`
	Instructions string `json:"description"`
	TravelTime   string `json:"traveltime"`
}
