package sync

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrNotPtr        = errors.New("target must be a pointer")
	ErrNotStruct     = errors.New("target must be a struct or a pointer to a struct")
	ErrFieldNotFound = errors.New("field not found in target struct")
	ErrTypeMismatch  = errors.New("type mismatch when applying delta")
)

func GenerateDelta(oldState, newState interface{}) ([]Delta, error) {
	oldVal := reflect.ValueOf(oldState)
	newVal := reflect.ValueOf(newState)

	if oldVal.Kind() == reflect.Ptr {
		oldVal = oldVal.Elem()
	}
	if newVal.Kind() == reflect.Ptr {
		newVal = newVal.Elem()
	}

	if oldVal.Kind() != reflect.Struct || newVal.Kind() != reflect.Struct {
		return nil, ErrNotStruct
	}
	if oldVal.Type() != newVal.Type() {
		return nil, fmt.Errorf("type mismatch: old=%s, new=%s", oldVal.Type(), newVal.Type())
	}

	deltas := make([]Delta, 0)
	structType := oldVal.Type()

	for i := 0; i < oldVal.NumField(); i++ {
		fieldName := structType.Field(i).Name

		if !oldVal.Field(i).CanInterface() {
			continue
		}

		oldFieldValue := oldVal.Field(i).Interface()
		newFieldValue := newVal.Field(i).Interface()

		if !reflect.DeepEqual(oldFieldValue, newFieldValue) {

			deltas = append(deltas, Delta{
				FieldName: fieldName,
				NewValue:  newFieldValue,
			})
		}
	}

	return deltas, nil
}

func ApplyDelta(targetStatePtr interface{}, deltas []Delta) error {
	targetVal := reflect.ValueOf(targetStatePtr)

	if targetVal.Kind() != reflect.Ptr {
		return ErrNotPtr
	}
	targetElem := targetVal.Elem()

	if targetElem.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	for _, delta := range deltas {

		fieldVal := targetElem.FieldByName(delta.FieldName)
		if !fieldVal.IsValid() {

			fmt.Printf("[Warning] ApplyDelta: Field '%s' not found in target struct %s\n", delta.FieldName, targetElem.Type())

			continue
		}

		if !fieldVal.CanSet() {
			fmt.Printf("[Warning] ApplyDelta: Field '%s' cannot be set in target struct %s\n", delta.FieldName, targetElem.Type())
			continue
		}

		newValue := reflect.ValueOf(delta.NewValue)

		if fieldVal.Type() != newValue.Type() {

			if newValue.CanConvert(fieldVal.Type()) {
				newValue = newValue.Convert(fieldVal.Type())
			} else {
				fmt.Printf("[Warning] ApplyDelta: Type mismatch for field '%s'. Expected %s, got %s\n", delta.FieldName, fieldVal.Type(), newValue.Type())

				continue
			}
		}

		fieldVal.Set(newValue)
	}

	return nil
}
