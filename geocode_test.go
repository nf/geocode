package geocode

import "testing"

func TestLookup(t *testing.T) {
	req := &Request{
		Address:  "New York City",
		Provider: GOOGLE,
	}
	resp, err := req.Lookup(nil)
	if err != nil {
		t.Fatalf("Lookup error: %v", err)
	}
	if s := resp.Status; s != "OK" {
		t.Fatalf(`Status == %q, want "OK"`, s)
	}
	if l := len(resp.Results); l != 1 {
		t.Fatalf("len(Results) == %d, want 1", l)
	}
	addr := "New York, NY, USA"
	if a := resp.Found; a != addr {
		t.Errorf("Address == %q, want %q", a, addr)
	}
}

func TestLookupWithBounds(t *testing.T) {
	req := &Request{
		Address:  "Winnetka",
		Provider: GOOGLE,
	}
	bounds := &Bounds{Point{34.172684, -118.604794},
		Point{34.236144, -118.500938}}
	req.Bounds = bounds
	resp, err := req.Lookup(nil)
	if err != nil {
		t.Fatalf("Lookup error: %v", err)
	}
	if s := resp.Status; s != "OK" {
		t.Fatalf(`Status == %q, want "OK"`, s)
	}
	if l := len(resp.Results); l != 1 {
		t.Fatalf("len(Results) == %d, want 1", l)
	}
	addr := "Winnetka, Los Angeles, CA, USA"
	if a := resp.Found; a != addr {
		t.Errorf("Address == %q, want %q", a, addr)
	}
}

func TestLookupWithLanguage(t *testing.T) {
	req := &Request{
		Address:  "札幌市",
		Provider: GOOGLE,
	}
	req.Language = "ja"
	resp, err := req.Lookup(nil)
	if err != nil {
		t.Fatalf("Lookup error: %v", err)
	}
	if s := resp.Status; s != "OK" {
		t.Fatalf(`Status == %q, want "OK"`, s)
	}
	if l := len(resp.Results); l != 1 {
		t.Fatalf("len(Results) == %d, want 1", l)
	}
	addr := "日本, 北海道札幌市"
	if a := resp.Found; a != addr {
		t.Errorf("Address == %q, want %q", a, addr)
	}
}

func TestLookupWithRegion(t *testing.T) {
	req := &Request{
		Address:  "Toledo",
		Provider: GOOGLE,
	}
	req.Region = "es"
	resp, err := req.Lookup(nil)
	if err != nil {
		t.Fatalf("Lookup error: %v", err)
	}
	if s := resp.Status; s != "OK" {
		t.Fatalf(`Status == %q, want "OK"`, s)
	}
	if l := len(resp.Results); l != 1 {
		t.Fatalf("len(Results) == %d, want 1", l)
	}
	addr := "Toledo, Spain"
	if a := resp.Found; a != addr {
		t.Errorf("Address == %q, want %q", a, addr)
	}
}
