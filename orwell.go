package orwell

import (
	"reflect"
)

type (
	//Rule interface
	Rule interface {
		Apply(v interface{}) error
	}

	// FieldableError interface
	FieldableError interface {
		error
		FieldName() string
		JSONName() string
	}

	// InternalError interface
	InternalError interface {
		error
		InternalError() error
	}

	// IterableError interface
	IterableError interface {
		error
		Len() int
		ValueAt(int) error
	}

	orwell struct{}

	fieldRules struct {
		fieldPtr interface{}
		rules    []Rule
	}
)

// NewOrwell func
func NewOrwell() *orwell {
	return &orwell{}
}

// Validate func
func (o *orwell) Validate(v interface{}, rules ...Rule) error {
	for _, r := range rules {
		if err := r.Apply(v); err != nil {
			return err
		}
	}

	return nil
}

// ValidateStruct func
func (o *orwell) ValidateStruct(structPtr interface{}, fieldRules ...*fieldRules) error {
	structElem := reflect.ValueOf(structPtr).Elem()
	var structValidationError structValidationError

	for _, fieldRule := range fieldRules {
		fieldValue := reflect.ValueOf(fieldRule.fieldPtr)
		fieldName, jsonName := field(structElem, fieldValue)
		if err := o.Validate(fieldValue.Elem().Interface(), fieldRule.rules...); err != nil {
			if ie, ok := err.(InternalError); ok {
				return ie
			}
			structValidationError.errors = append(structValidationError.errors, NewValidationError(fieldName, jsonName, err.Error()))
		}
	}

	if len(structValidationError.errors) > 0 {
		return &structValidationError
	}

	return nil
}

// FieldRules func
func (o *orwell) FieldRules(field interface{}, rules ...Rule) *fieldRules {
	return &fieldRules{
		fieldPtr: field,
		rules:    rules,
	}
}

func field(structElem reflect.Value, fieldValue reflect.Value) (string, string) {
	var fieldName string
	var jsonName string
	fieldPointer := fieldValue.Pointer()
	for i := 0; i < structElem.NumField(); i++ {
		structField := structElem.Type().Field(i)
		if (fieldPointer == structElem.Field(i).UnsafeAddr()) && (structField.Type == fieldValue.Elem().Type()) {
			fieldName = structField.Name
			if str, ok := structField.Tag.Lookup("json"); ok {
				jsonName = str
			}
			return fieldName, jsonName
		}
		switch structElem.Field(i).Kind() {
		case reflect.Struct:
			return field(structElem.Field(i), fieldValue)
		}
	}

	return fieldName, jsonName
}