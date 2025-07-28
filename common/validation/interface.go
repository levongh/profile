package validation

import (
    "reflect"
    "strings"
)

type Validatable interface {
    Validate() *Result // Validate must ALWAYS be declared on a pointer to a struct
}

// Validate clean data (strips spaces from strings) and validates
// Make sure to provide a pointer to the struct or else it will fail
func Validate(v Validatable) *Result {
    trimSpaces(v)
    return v.Validate()
}

// will trim spaces from all string and *string fields of a struct
func trimSpaces(i interface{}) {
    v := reflect.ValueOf(i)

    // we need only pointers to make sure we changes will be seen outside
    if v.Kind() != reflect.Ptr {
        return // nothing will happen
    }

    // need to dereference pointer to access fields
    v = v.Elem()

    for i := 0; i < v.NumField(); i++ {
        str, ok := v.Field(i).Interface().(string)
        if ok {
            // we can set if it is a simple string
            str = strings.TrimSpace(str)
            v.Field(i).SetString(str)
        }

        ptr, ok := v.Field(i).Interface().(*string)
        if ok && ptr != nil {
            // if this is a pointer we set it like this
            *ptr = strings.TrimSpace(*ptr)
            v.Field(i).Set(reflect.ValueOf(ptr))
        }
    }
}