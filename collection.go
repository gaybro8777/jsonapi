package jsonapi

// A Collection defines the interface of a structure that can manage a set of
// ordered resources of the same type.
type Collection interface {
	// Type returns the name of the resources' type.
	GetType() Type

	// Len returns the number of resources in the collection.
	Len() int

	// At returns the resource at index i.
	At(int) Resource

	// Add adds a resource in the collection.
	Add(Resource)
}

// Resources is a slice of objects that implement the Resource interface. They
// do not necessarily have the same type.
type Resources []Resource

// GetType returns a zero Type object because the collection does not represent
// a particular type.
func (r *Resources) GetType() Type {
	return Type{}
}

// GetType returns the number of elements in r.
func (r *Resources) Len() int {
	return len(*r)
}

// GetType returns the number of elements in r.
func (r *Resources) At(i int) Resource {
	if i >= 0 && i < r.Len() {
		return (*r)[i]
	}
	return nil
}

// GetType adds a Resource object to r.
func (r *Resources) Add(res Resource) {
	*r = append(*r, res)
}
