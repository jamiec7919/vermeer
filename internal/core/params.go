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
var ErrParamNotBool = errors.New("Parameter is not bool or convertible to bool")
var ErrParamNotString = errors.New("Parameter is not string")
var ErrNotStruct = errors.New("Unmarshal: param is not a struct")
var ErrNotFound = errors.New("Parameter not found")

type Params map[string]interface{}

func (np *Params) GetString(key string) (string, error) {
	if v, present := (*np)[key]; present {
		return getString(v)
	}

	return "", ErrNotFound
}

func (np *Params) GetBool(key string) (bool, error) {
	if v, present := (*np)[key]; present {
		switch t := v.(type) {
		case float64:
			return t > 0.0, nil
		case int64:
			return t > 0, nil
		default:
			return false, ErrParamNotBool
		}
	}

	return false, ErrNotFound
}

func (np *Params) GetFloat(key string) (float64, error) {
	if v, present := (*np)[key]; present {
		return getFloat(v)
	}

	return 0.0, ErrNotFound
}

func (np *Params) GetInt(key string) (int64, error) {
	if v, present := (*np)[key]; present {
		return getInt(v)
	}

	return 0, ErrNotFound
}

func (np *Params) GetVec3(key string) (m.Vec3, error) {
	if v, present := (*np)[key]; present {
		return getVec3(v)
	}

	return m.Vec3{}, ErrNotFound
}

func (np *Params) GetVec2(key string) (m.Vec2, error) {
	if v, present := (*np)[key]; present {
		return getVec2(v)
	}

	return m.Vec2{}, ErrNotFound
}

func (np *Params) GetMatrix4(key string) (m.Matrix4, error) {
	if v, present := (*np)[key]; present {
		switch t := v.(type) {
		case m.Matrix4:
			return t, nil
		}
	}

	return m.Matrix4{}, ErrNotFound
}

func (np *Params) GetVec3Slice(key string) ([]m.Vec3, error) {
	if v, present := (*np)[key]; present {
		switch t := v.(type) {
		case []m.Vec3:
			return t, nil
		}
	}

	return nil, ErrNotFound
}

func (np *Params) GetInt32Slice(key string) ([]int32, error) {
	if v, present := (*np)[key]; present {
		switch t := v.(type) {
		case []int32:
			return t, nil
		}
	}

	return nil, ErrNotFound
}

func (np *Params) GetVec2Slice(key string) ([]m.Vec2, error) {
	if v, present := (*np)[key]; present {
		switch t := v.(type) {
		case []m.Vec2:
			return t, nil
		}
	}

	return nil, ErrNotFound
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

func getVec2(v interface{}) (m.Vec2, error) {
	switch t := v.(type) {
	case []float64:
		if len(t) == 2 {
			return m.Vec2{float32(t[0]), float32(t[1])}, nil
		} else {
			return m.Vec2{}, nil
		}
	case []int64:
		if len(t) == 2 {
			return m.Vec2{float32(t[0]), float32(t[1])}, nil
		} else {
			return m.Vec2{}, nil
		}
	default:
		return m.Vec2{}, nil
	}
}

// This is a utility function for simple structs representing nodes
func (p *Params) Unmarshal(v interface{}) error {
	// Check v is struct

	rv := reflect.ValueOf(v)

	relem := rv.Elem()

	if relem.Kind() != reflect.Struct {
		return ErrNotStruct

	}

	ty := relem.Type()

	for i := 0; i < relem.NumField(); i++ {
		f := ty.Field(i)

		fieldName := f.Name

		if tag := f.Tag.Get("node"); tag != "" {
			fieldName = tag
		}

		fv := relem.Field(i)

		if fv.CanSet() {
			switch fv.Kind() {
			case reflect.Int:
				v, err := p.GetInt(fieldName)
				if err == nil {
					fv.SetInt(v)
				}
			case reflect.Float64:
				fallthrough
			case reflect.Float32:
				v, err := p.GetFloat(fieldName)
				if err == nil {
					fv.SetFloat(v)
				}
			case reflect.String:
				v, err := p.GetString(fieldName)
				if err == nil {
					fv.SetString(v)
				}
			case reflect.Bool:
				v, err := p.GetBool(fieldName)
				if err == nil {
					fv.SetBool(v)
				}

			case reflect.Slice:
				switch fv.Type().Elem() {
				case reflect.TypeOf(m.Vec3{}):
					v, err := p.GetVec3Slice(fieldName)
					if err == nil {
						fv.Set(reflect.ValueOf(v))
					}
				case reflect.TypeOf(int32(0)):

					v, err := p.GetInt32Slice(fieldName)
					if err == nil {
						fv.Set(reflect.ValueOf(v))
					}
				case reflect.TypeOf(m.Vec2{}):
					v, err := p.GetVec2Slice(fieldName)
					if err == nil {
						fv.Set(reflect.ValueOf(v))
					}
				}

			default:
				switch fv.Type() {
				case reflect.TypeOf(m.Vec3{}):
					v, err := p.GetVec3(fieldName)
					if err == nil {
						fv.Set(reflect.ValueOf(v))
					}
				case reflect.TypeOf(m.Vec2{}):
					v, err := p.GetVec2(fieldName)
					if err == nil {
						fv.Set(reflect.ValueOf(v))
					}
				case reflect.TypeOf(m.Matrix4{}):
					v, err := p.GetMatrix4(fieldName)
					if err == nil {
						fv.Set(reflect.ValueOf(v))
					}
				}
			}

		}
	}

	return nil
}
