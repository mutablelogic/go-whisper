package httprequest

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

/////////////////////////////////////////////////////////////////////////////
// TYPES

/////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	tagName = "json"
)

var (
	typeTime          = reflect.TypeOf(time.Time{})
	typeFileHeaderPtr = reflect.TypeOf((*multipart.FileHeader)(nil))
)

/////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ReadQuery reads the query parameters from a URL and sets the fields of a
// struct v with the values. The struct v must be a pointer to a struct.
func ReadQuery(v any, q url.Values) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errors.New("v must be a pointer")
	} else {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return errors.New("v must be a pointer to a struct")
	}

	// Enumerate fields
	fields := reflect.VisibleFields(rv.Type())
	if len(fields) == 0 {
		return errors.New("v has no public fields")
	}
	for _, field := range fields {
		tag := jsonName(field)
		if tag == "" {
			continue
		}
		v := rv.FieldByName(field.Name)
		if !v.CanSet() {
			continue
		}
		if value, exists := q[tag]; exists {
			if err := setQueryValue(tag, rv.FieldByIndex(field.Index), value); err != nil {
				return err
			}
		} else {
			if err := setQueryValue(tag, rv.FieldByIndex(field.Index), nil); err != nil {
				return err
			}
		}
	}

	// Return success
	return nil
}

/////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// readFiles reads file parameters from a form and sets the fields of a
// struct v with the file handles. The struct v must be a pointer to a struct.
// and file fields must be of type *multipart.FileHeader or []*multipart.FileHeader.
func readFiles(v any, q map[string][]*multipart.FileHeader) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errors.New("v must be a pointer")
	} else {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return errors.New("v must be a pointer to a struct")
	}

	// Enumerate fields
	fields := reflect.VisibleFields(rv.Type())
	if len(fields) == 0 {
		return errors.New("v has no public fields")
	}
	for _, field := range fields {
		tag := jsonName(field)
		if tag == "" {
			continue
		}
		v := rv.FieldByName(field.Name)
		if !v.CanSet() {
			continue
		}
		if value, exists := q[tag]; exists {
			if err := setFileValue(tag, rv.FieldByIndex(field.Index), value); err != nil {
				return err
			}
		}
	}

	// Return success
	return nil
}

func jsonName(field reflect.StructField) string {
	tag := field.Tag.Get(tagName)
	if tag == "-" {
		return ""
	}
	if fields := strings.Split(tag, ","); len(fields) > 0 && fields[0] != "" {
		return fields[0]
	}
	return field.Name
}

func setQueryValue(tag string, v reflect.Value, value []string) error {
	// Set zero-value
	if len(value) == 0 {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	// Create a new value for pointers
	if v.Kind() == reflect.Ptr {
		v.Set(reflect.New(v.Type().Elem()))
		v = v.Elem()
	}
	// Set the value
	switch v.Kind() {
	case reflect.String:
		if len(value) > 0 {
			v.SetString(value[0])
		}
	case reflect.Bool:
		value, err := strconv.ParseBool(value[0])
		if err != nil {
			return fmt.Errorf("%q: %w", tag, err)
		}
		v.SetBool(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := strconv.ParseInt(value[0], 10, 64)
		if err != nil {
			return fmt.Errorf("%q: %w", tag, err)
		}
		v.SetInt(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value, err := strconv.ParseUint(value[0], 10, 64)
		if err != nil {
			return fmt.Errorf("%q: %w", tag, err)
		}
		v.SetUint(value)
	case reflect.Struct:
		// TODO: Abstract this to unmarshaler interface with UnmarshalJSON
		// from a string-quoted value
		switch v.Type() {
		case typeTime:
			t := new(time.Time)
			if len(value) > 0 {
				quoted := strconv.Quote(value[0])
				if err := t.UnmarshalJSON([]byte(quoted)); err != nil {
					return fmt.Errorf("%q: %w", tag, err)
				}
			}
			v.Set(reflect.ValueOf(t).Elem())
		default:
			return fmt.Errorf("%q: unsupported type (%q)", tag, v.Type())
		}
	default:
		return fmt.Errorf("%q: unsupported kind (%q)", tag, v.Kind())
	}

	// Return success
	return nil
}

func setFileValue(tag string, v reflect.Value, value []*multipart.FileHeader) error {
	// Set zero-value
	if len(value) == 0 {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	// Set the value
	t := v.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
		switch v.Type() {
		case typeFileHeaderPtr:
			v.Set(reflect.ValueOf(value[0]))
		default:
			return fmt.Errorf("%q: unsupported struct type (%q)", tag, v.Type())
		}
	case reflect.Slice:
		switch v.Type().Elem() {
		case typeFileHeaderPtr:
			slice := reflect.MakeSlice(v.Type(), len(value), len(value))
			for i, file := range value {
				slice.Index(i).Set(reflect.ValueOf(file))
			}
			v.Set(slice)
		default:
			return fmt.Errorf("%q: unsupported kind (%q)", tag, v.Kind())
		}
	default:
		return fmt.Errorf("%q: unsupported kind (%q)", tag, v.Kind())
	}

	// Return success
	return nil
}
