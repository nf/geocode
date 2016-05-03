geocode
=======

Go package for interacting with:
- Google's Geocoding API
- OSM's Nomatin API (To be more precise: MapQuestApi's Geocoding API)
- YOUR's routing API

For Google's GEOCODING:

```
	req := &geocode.Request{
		Region:   "us",
		Provider: geocode.GOOGLE,
		Key: "YOUR_GOOGLE_MAPS_API_KEY",
		Location: &geocode.Point{34.64973, -98.41503},
  }
```

For OSM/MAPQUEST GEOCODING:
```
	req := &geocode.Request{
		Provider: geocode.OSM,
		Limit:    1,
		Region: "ar",
		Key: "YOUR_MAPQUEST_API_KEY",
		Location: &geocode.Point{34.64973, -98.41503},
	}
```

Let's say you need to reverse geocode locations, and want to rely on all providers. First we hit OSM/MAPQUEST, then fall back to GOOGLE:

```
	req := &geocode.Request{
		Limit:  1,
		Type:   geocode.GEOCODE,
		Region: "ar",
		Location: &geocode.Point{-34.649966, -58.421769},
	}

	sucessfulLookup := false

	req.Provider = geocode.OSM
	req.Key = "YOUR_MAPQUEST_API_KEY"

	resp, err := req.Lookup(nil)
	if err != nil {
		// fmt.Printf("Lookup error: %v - %v \n", resp, err)
	} else {
		// fmt.Printf("----> OSM query string is: %s \n", resp.QueryString)
		if s := resp.Status; s != "OK" {
			fmt.Printf(`%s: sStatus == %q \n`, req.Location, s)
		} else {
			if resp.Count > 0 {
				sucessfulLookup = true
			}
		}
	}

	if !sucessfulLookup {
		req.Provider = geocode.GOOGLE
		req.Key = "YOUR_GOOGLE_MAPS_API_KEY"

		resp, err = req.Lookup(nil)
		if err != nil {
			// fmt.Printf("Lookup error: %v\n", err)
		} else {
			// fmt.Printf("----> Google query string is: %s \n", resp.QueryString)
			if s := resp.Status; s != "OK" {
				fmt.Printf(`%s: sStatus == %q \n`, req.Location, s)
			} else {
				sucessfulLookup = true
			}
		}

	}

	if sucessfulLookup {
		fmt.Printf("%s \n", resp.Found)
		time.Sleep(time.Duration(rand.Intn(200-1)+1) * time.Millisecond)
	}
```

Additionally, you could do routing. 

For YOURS ROUTING:
```
	req := &geocode.Request{
		Provider: geocode.YOURS,
		Bounds: &geocode.Bounds{geocode.Point{34.172684, -118.604794},
			geocode.Point{34.236144, -118.500938}}
	}
```

Then:
```
  resp, err := req.Route(nil)
  if err != nil {
		continue
  } else {
    if s := resp.Status; s != "OK" {
		continue
    } else {
		fmt.Printf("---> Distance: %s, Instructions: %s\n", resp.YOURSResponse.Properties.Distance, resp.YOURSResponse.Properties.Instructions)
		for k := range resp.YOURSResponse.Coordinates {
			fmt.Printf("\t[%d] %f, %f\n", k, resp.YOURSResponse.Coordinates[k][0], resp.YOURSResponse.Coordinates[k][1])
		}
    }
  }
```

If using Google App Engine, you'll need to set the geocode request's HTTPClient:

```
	import (
		"appengine"
		"appengine/urlfetch"
	)

	c := appengine.NewContext(req)

	geoReq := &geocode.Request{
		HTTPClient: urlfetch.Client(c),
		// ...
	}
```
