// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"errors"
	m "github.com/jamiec7919/vermeer/math"
	"reflect"
)

var ErrParamNotFloat = errors.New("Parameter is not float or convertible to float")
var ErrParamNotInt = errors.New("Parameter is not int or convertible to int")
var ErrParamNotString = errors.New("Parameter is not string")
var ErrNotStruct = errors.New("Unmarshal: param is not a struct")

type Params map[string]interface{}

func (np *Params) GetString(key string) (string, error) {
	if v, present := (*np)[key]; present {
		return getString(v)
	}

	return "", nil
}

func (np *Params) GetFloat(key string) (float64, error) {
	if v, present := (*np)[key]; present {
		return getFloat(v)
	}

	return 0.0, nil
}

func (np *Params) GetInt(key string) (int64, error) {
	if v, present := (*np)[key]; present {
		return getInt(v)
	}

	return 0, nil
}

func (np *Params) GetVec3(key string) (m.Vec3, error) {
	if v, present := (*np)[key]; present {
		return getVec3(v)
	}

	return m.Vec3{}, nil
}

func getString(v interface{}) (string, error) {
	switch t := v.(type) {
	case string:
		return t, nil
	default:
		return "", ErrParamNotString
	}

}

func getFloat(v interface{}) (float64, error) {
	switch t := v.(type) {
	case float64:
		return t, nil
	case int64:
		return float64(t), nil
	default:
		return 0, ErrParamNotFloat
	}

}

func getInt(v interface{}) (int64, error) {
	switch t := v.(type) {
	case float64:
		return int64(t), nil
	case int64:
		return t, nil
	default:
		return 0, ErrParamNotInt
	}

}

func getVec3(v interface{}) (m.Vec3, error) {
	switch t := v.(type) {
	case []float64:
		if len(t) == 3 {
			return m.Vec3{float32(t[0]), float32(t[1]), float32(t[2])}, nil
		} else {
			return m.Vec3{}, nil
		}
	case []int64:
		if len(t) == 3 {
			return m.Vec3{float32(t[0]), float32(t[1]), float32(t[2])}, nil
		} else {
			return m.Vec3{}, nil
		}
	default:
		return m.Vec3{}, nil
	}
}

// This is a utility function for simple structs representing nodes
func (np *Params) Unmarshal(v interface{}) error {
	// Check v is struct

	rv := reflect.ValueOf(v)

	relem := rv.Elem()

	if relem.Kind() != reflect.Struct {
		return ErrNotStruct

	}

	for k, param := range *np {
		// lookup k in v and assign as appropriate
		field := relem.FieldByName(k)
		// log.Printf("%v", k)
		if field.IsValid() {
			if field.CanSet() {
				switch field.Kind() {
				case reflect.Int:
					v, _ := getInt(param)
					field.SetInt(v)
				case reflect.Float64:
					fallthrough
				case reflect.Float32:
					v, _ := getFloat(param)
					field.SetFloat(v)
				case reflect.String:
					field.SetString(param.(string))
				case reflect.Slice:
					// el := field.Type().Elem()

					arr := param.([]Params)

					v := reflect.MakeSlice(field.Type(), len(arr), len(arr))

					for i := range arr {
						arr[i].Unmarshal(v.Index(i).Addr().Interface())
						// log.Printf("%v %v", i, arr[i])
					}
					field.Set(v)
					// log.Printf("Filling array of %v %v", el, v)
				default:
					switch field.Type() {
					case reflect.TypeOf(m.Vec3{}):
						v, _ := getVec3(param)
						field.Set(reflect.ValueOf(v))
						//				case reflect.TypeOf(Colour{}):
						//					v, _ := getColour(param)
						//					field.Set(reflect.ValueOf(v))
						//				case reflect.TypeOf((*Texture)(nil)):
						//					v, _ := getTexture(param)
						//					field.Set(reflect.ValueOf(v))
					}
				}
			}
		}
	}

	return nil

}
