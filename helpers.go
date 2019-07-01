package jsonapi

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// CheckType checks the given value and returns any error found.
//
// If nil is returned, than the value can be safely used with this library.
func CheckType(v interface{}) error {
	value := reflect.ValueOf(v)
	kind := value.Kind()

	// Check wether it's a struct
	if kind != reflect.Struct {
		return errors.New("jsonapi: not a struct")
	}

	// Check ID field
	var (
		idField reflect.StructField
		ok      bool
	)
	if idField, ok = value.Type().FieldByName("ID"); !ok {
		return errors.New("jsonapi: struct doesn't have an ID field")
	}

	resType := idField.Tag.Get("api")
	if resType == "" {
		return errors.New("jsonapi: ID field's api tag is empty")
	}

	// Check attributes
	for i := 0; i < value.NumField(); i++ {
		sf := value.Type().Field(i)

		if sf.Tag.Get("api") == "attr" {
			isValid := false

			switch sf.Type.String() {
			case "string", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "bool", "time.Time", "*string", "*int", "*int8", "*int16", "*int32", "*int64", "*uint", "*uint8", "*uint16", "*uint32", "*uint64", "*bool", "*time.Time":
				isValid = true
			}

			if !isValid {
				return fmt.Errorf("jsonapi: attribute %s of type %s is of unsupported type", sf.Name, resType)
			}
		}
	}

	// Check relationships
	for i := 0; i < value.NumField(); i++ {
		sf := value.Type().Field(i)

		if strings.HasPrefix(sf.Tag.Get("api"), "rel,") {
			s := strings.Split(sf.Tag.Get("api"), ",")

			if len(s) < 2 || len(s) > 3 {
				return fmt.Errorf("jsonapi: api tag of relationship %s of struct %s is invalid", sf.Name, value.Type().Name())
			}

			if sf.Type.String() != "string" && sf.Type.String() != "[]string" {
				return fmt.Errorf("jsonapi: relationship %s of type %s is not string or []string", sf.Name, resType)
			}
		}
	}

	return nil
}

// IDAndType returns the ID and the type of the resource represented by v.
//
// Two empty strings are returned if v is not recognized as a resource.
// CheckType can be used to check the validity of a struct.
func IDAndType(v interface{}) (string, string) {
	switch nv := v.(type) {
	case Resource:
		return nv.GetID(), nv.GetType().Name
	}

	val := reflect.ValueOf(v)

	// Allows pointers to structs
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		idF := val.FieldByName("ID")

		if !idF.IsValid() {
			return "", ""
		}

		idSF, _ := val.Type().FieldByName("ID")

		if idF.Kind() == reflect.String {
			return idF.String(), idSF.Tag.Get("api")
		}
	}

	return "", ""
}
