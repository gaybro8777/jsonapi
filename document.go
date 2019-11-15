package jsonapi

import (
	"encoding/json"
	"errors"
	"sort"
)

// A Document represents a JSON:API document.
type Document struct {
	// Data
	Data interface{}

	// Included
	Included []Resource

	// References
	Resources map[string]map[string]struct{}
	Links     map[string]Link

	// Relationships where data has to be included in payload
	RelData map[string][]string

	// Top-level members
	Meta map[string]interface{}

	// Errors
	Errors []Error

	// Internal
	PrePath string
}

// Include adds res to the set of resources to be included under the included
// top-level field.
//
// It also makes sure that resources are not added twice.
func (d *Document) Include(res Resource) {
	key := res.GetID() + " " + res.GetType().Name

	if len(d.Included) == 0 {
		d.Included = []Resource{}
	}

	if dres, ok := d.Data.(Resource); ok {
		// Check resource
		rkey := dres.GetID() + " " + dres.GetType().Name

		if rkey == key {
			return
		}
	} else if col, ok := d.Data.(Collection); ok {
		// Check Collection
		ctyp := col.GetType()
		if ctyp.Name == res.GetType().Name {
			for i := 0; i < col.Len(); i++ {
				rkey := col.At(i).GetID() + " " + col.At(i).GetType().Name

				if rkey == key {
					return
				}
			}
		}
	}

	// Check already included resources
	for _, res := range d.Included {
		if key == res.GetID()+" "+res.GetType().Name {
			return
		}
	}

	d.Included = append(d.Included, res)
}

// MarshalDocument marshals a document according to the JSON:API speficication.
//
// Both doc and url must not be nil.
func MarshalDocument(doc *Document, url *URL) ([]byte, error) {
	var err error

	// Data
	var data json.RawMessage
	if res, ok := doc.Data.(Resource); ok {
		// Resource
		data = MarshalResource(
			res,
			doc.PrePath,
			url.Params.Fields[res.GetType().Name],
			doc.RelData,
		)
	} else if col, ok := doc.Data.(Collection); ok {
		// Collection
		data = MarshalCollection(
			col,
			doc.PrePath,
			url.Params.Fields,
			doc.RelData,
		)
	} else if id, ok := doc.Data.(Identifier); ok {
		// Identifier
		data, err = json.Marshal(id)
	} else if ids, ok := doc.Data.(Identifiers); ok {
		// Identifiers
		data, err = json.Marshal(ids)
	} else if doc.Data != nil {
		err = errors.New("data contains an unknown type")
	} else if len(doc.Errors) == 0 {
		data = []byte("null")
	}

	// Data
	var errors json.RawMessage
	if len(doc.Errors) > 0 {
		// Errors
		errors, err = json.Marshal(doc.Errors)
	}

	if err != nil {
		return []byte{}, err
	}

	// Included
	var inclusions []*json.RawMessage

	if len(doc.Included) > 0 {
		sort.Slice(doc.Included, func(i, j int) bool {
			return doc.Included[i].GetID() < doc.Included[j].GetID()
		})

		if len(data) > 0 {
			for key := range doc.Included {
				typ := doc.Included[key].GetType().Name
				raw := MarshalResource(
					doc.Included[key],
					doc.PrePath,
					url.Params.Fields[typ],
					doc.RelData,
				)
				rawm := json.RawMessage(raw)
				inclusions = append(inclusions, &rawm)
			}
		}
	}

	// Marshaling
	plMap := map[string]interface{}{}

	if len(errors) > 0 {
		plMap["errors"] = errors
	} else if len(data) > 0 {
		plMap["data"] = data

		if len(inclusions) > 0 {
			plMap["included"] = inclusions
		}
	}

	if len(doc.Meta) > 0 {
		plMap["meta"] = doc.Meta
	}

	if url != nil {
		plMap["links"] = map[string]string{
			"self": doc.PrePath + url.String(),
		}
	}

	plMap["jsonapi"] = map[string]string{"version": "1.0"}

	return json.Marshal(plMap)
}

// UnmarshalDocument reads a payload to build and return a Document object.
//
// schema must not be nil.
func UnmarshalDocument(payload []byte, schema *Schema) (*Document, error) {
	doc := &Document{
		Included:  []Resource{},
		Resources: map[string]map[string]struct{}{},
		Links:     map[string]Link{},
		RelData:   map[string][]string{},
		Meta:      map[string]interface{}{},
	}
	ske := &payloadSkeleton{}

	// Unmarshal
	err := json.Unmarshal(payload, ske)
	if err != nil {
		return nil, err
	}

	// Data
	if len(ske.Data) > 0 {
		if ske.Data[0] == '{' {
			// Resource
			res, err := UnmarshalResource(ske.Data, schema)
			if err != nil {
				return nil, err
			}

			doc.Data = res
		} else if ske.Data[0] == '[' {
			col, err := UnmarshalCollection(ske.Data, schema)
			if err != nil {
				return nil, err
			}
			doc.Data = col
		} else if string(ske.Data) == "null" {
			doc.Data = nil
		} else {
			// TODO Not exactly the right error
			return nil, NewErrMissingDataMember()
		}
	} else if len(ske.Errors) > 0 {
		doc.Errors = ske.Errors
	} else {
		return nil, NewErrMissingDataMember()
	}

	// Included
	if len(ske.Included) > 0 {
		inc := Identifier{}
		incs := []Identifier{}

		for _, rawInc := range ske.Included {
			err = json.Unmarshal(rawInc, &inc)
			if err != nil {
				return nil, err
			}

			incs = append(incs, inc)
		}

		for i := range incs {
			res, err := UnmarshalResource(ske.Included[i], schema)
			if err != nil {
				return nil, err
			}

			doc.Included = append(doc.Included, res)
		}
	}

	// Meta
	doc.Meta = ske.Meta

	return doc, nil
}
