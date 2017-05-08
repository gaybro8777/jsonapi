package jsonapi

import (
	"encoding/json"
)

// Document ...
type Document struct {
	// Data
	Resource    Resource
	Collection  Collection
	Identifier  Identifier
	Identifiers Identifiers

	// Included
	Included map[string]Resource

	// References
	Resources map[string]map[string]struct{}
	Links     map[string]Link

	// Options
	Options *Options

	// Errors
	Errors []Error

	// URL
	URL *URL
}

// NewDocument ...
func NewDocument() *Document {
	return &Document{
		Included:  map[string]Resource{},
		Resources: map[string]map[string]struct{}{},
		Links:     map[string]Link{},
		Options:   NewOptions("", nil),
		URL:       NewURL(),
	}
}

// Include ...
func (d *Document) Include(res Resource) {
	id, typ := res.IDAndType()
	key := typ + " " + id

	if len(d.Included) == 0 {
		d.Included = map[string]Resource{}
	}

	// Check resource
	if d.Resource != nil {
		rid, rtype := d.Resource.IDAndType()
		rkey := rid + " " + rtype

		if rkey == key {
			return
		}
	}

	// Check Collection
	if d.Collection != nil {
		_, ctyp := d.Collection.Sample().IDAndType()
		if ctyp == typ {
			for i := 0; i < d.Collection.Len(); i++ {
				rid, rtype := d.Collection.Elem(i).IDAndType()
				rkey := rid + " " + rtype

				if rkey == key {
					return
				}
			}
		}
	}

	// Check already included resources
	if _, ok := d.Included[key]; ok {
		return
	}

	d.Included[key] = res
}

// MarshalJSON ...
func (d *Document) MarshalJSON() ([]byte, error) {
	// Data
	var data json.RawMessage
	var errors json.RawMessage
	var err error
	if d.Errors != nil {
		errors, err = json.Marshal(d.Errors)
	} else if d.Resource != nil {
		data, err = d.Resource.MarshalJSONOptions(d.Options)
	} else if d.Collection != nil {
		data, err = d.Collection.MarshalJSONOptions(d.Options)
	} else if (d.Identifier != Identifier{}) {
		data, err = json.Marshal(d.Identifier)
	} else if d.Identifiers != nil {
		data, err = json.Marshal(d.Identifiers)
	} else {
		data = []byte("null")
	}

	if err != nil {
		return []byte{}, err
	}

	// Included
	inclusions := []*json.RawMessage{}
	if len(data) > 0 {
		for key := range d.Included {
			raw, err := d.Included[key].MarshalJSONOptions(d.Options)
			if err != nil {
				return []byte{}, err
			}
			rawm := json.RawMessage(raw)
			inclusions = append(inclusions, &rawm)
		}
	}

	// Marshaling
	plMap := map[string]interface{}{}

	if len(data) > 0 {
		plMap["data"] = data
	}

	if len(d.Links) > 0 {
		plMap["links"] = d.Links
	}

	if len(errors) > 0 {
		plMap["errors"] = errors
	}

	if len(inclusions) > 0 {
		plMap["included"] = inclusions
	}

	if len(d.Options.Meta) > 0 {
		plMap["meta"] = d.Options.Meta
	}

	if len(d.Options.JSONAPI) > 0 {
		plMap["jsonapi"] = d.Options.JSONAPI
	}

	return json.Marshal(plMap)
}
