package geocode

import "testing"

func TestLookup(t *testing.T) {
	req := &Request{
		Address: "New York City",
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
	if a := resp.Results[0].Address; a != addr {
		t.Errorf("Address == %q, want %q", a, addr)
	}
}
