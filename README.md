geocode
=======

Go package for interacting with:
- Google's Geocoding API
- OSM's Nomatin API
- YOUR's routing API

For Google's GEOCODING:

```
	req := &geocode.Request{
		Region:   "us",
		Provider: geocode.GOOGLE,
		Location: &geocode.Point{34.64973, -98.41503}
  }
```

For OSM GEOCODING:
```
	req := &geocode.Request{
		Provider: geocode.OSM,
		Limit:    1,
		Location: &geocode.Point{34.64973, -98.41503}
	}
```

Then:
```
  resp, err := req.Lookup(nil)
  if err != nil {
		//fmt.Printf("Lookup error: %v\n", err)
		continue
  } else {
    if s := resp.Status; s != "OK" {
			continue
			//fmt.Printf(`%s: sStatus == %q\n`, req.Location, s)
    } else {
			fmt.Printf("---> result[%d]: %s\n", resp.Count, resp.Found)
    }
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
