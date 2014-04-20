geocode
=======

Go package for interacting with both Google's Geocoding API, as well as OSM's Nomatin API. 

Forked from: https://github.com/nf/geocode

For Google:

```
	req := &geocode.Request{
		Region:   "us",
		Provider: geocode.GOOGLE,
  }
```

For OSM:
```
	req := &geocode.Request{
		Provider: geocode.OSM,
		Limit:    1,
	}
```

Then:
```
  req.Location = &geocode.Point{34.64973, -98.41503}
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
