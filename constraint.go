package gcfg

import (
	"fmt"
	"reflect"
	"github.com/baobabus/gcfg/types"
)

type constraints struct {
	min     string
	max     string
	minlen  int
	maxlen  int
}

type boundaryGetter func(interface{}, string) (*reflect.Value, error)

// Gets boundary value by scanning
func scanBoundary(d interface{}, val string) (*reflect.Value, error) {
	if val == "" { return nil, nil; }
	v := reflect.New(reflect.ValueOf(d).Elem().Type())
	r := &v
	if err := types.ScanFully(r.Interface(), val, 'v'); err != nil { return nil, err; }
	return r, nil
}

// Gets boundary value by calling UnmarshalText()
func unmarshalBoundary(d interface{}, val string) (*reflect.Value, error) {
	if val == "" { return nil, nil; }
	v := reflect.New(reflect.ValueOf(d).Elem().Type())
	r := &v
	dtu, ok := r.Interface().(textUnmarshaler)
	if !ok { return nil, errUnsupportedType; }
	if err := dtu.UnmarshalText([]byte(val)); err != nil { return nil, err; }
	return r, nil
}

// Checks whether d is within bounds specified in the metadata.
// This is only applicable to ordered types.
func checkBounds(d interface{}, t metadata, bg boundaryGetter) error {
	var obl, obh bool
	var vs, ls, us string
	min, err := bg(d, t.constraints.min); if err != nil { return err; }
	max, err := bg(d, t.constraints.max); if err != nil { return err; }
	if min != nil || max != nil {
		// Hack aimed specifically at time.Time for now
		// TODO add check for NumIn() and NumOut() and assert in and out types
		var z reflect.Value
		s := reflect.ValueOf(d).Elem()
		lm := s.MethodByName("Less"); if lm == z { lm = s.MethodByName("Before"); }
		gm := s.MethodByName("Greater"); if gm == z { gm = s.MethodByName("After"); }
		if lm != z && gm != z { // via explicit methods for ordering detection
			obl = min != nil && lm.Call([]reflect.Value {min.Elem()})[0].Bool()
			obh = max != nil && gm.Call([]reflect.Value {max.Elem()})[0].Bool()
			if min != nil { ls = fmt.Sprintf("%v", min.Interface()); }
			if max != nil { us = fmt.Sprintf("%v", max.Interface()); }
			vs = fmt.Sprintf("%v", s.Interface())
		} else { // via kind ordering
			rv := reflect.ValueOf(d)
			vk := rv.Type().Kind()
			v := reflect.ValueOf(d).Elem().Interface()
			if vk == reflect.Ptr {
				rv = rv.Elem()
				vk = rv.Type().Kind()
				if min != nil { v := min.Elem(); min = &v; }
				if max != nil { v := max.Elem(); max = &v; }
			}
			if min != nil { ls = fmt.Sprintf("%v", min.Interface()); }
			if max != nil { us = fmt.Sprintf("%v", max.Interface()); }
			vs = fmt.Sprintf("%v", v)
			switch vk {
			case reflect.String:
				if min != nil { ls = fmt.Sprintf("\"%s\"", ls); }
				if max != nil { us = fmt.Sprintf("\"%s\"", us); }
				obl = min != nil && rv.String() < min.String()
				obh = max != nil && rv.String() > max.String()
				vs = fmt.Sprintf("\"%s\"", vs)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				obl = min != nil && rv.Int() < min.Int()
				obh = max != nil && rv.Int() > max.Int()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				obl = min != nil && rv.Uint() < min.Uint()
				obh = max != nil && rv.Uint() > max.Uint()
			case reflect.Float32, reflect.Float64:
				obl = min != nil && rv.Float() < min.Float()
				obh = max != nil && rv.Float() > max.Float()
			}
		}
	}
	if obl || obh {
		if min != nil && max != nil {
			return fmt.Errorf("Value %s out of bounds [%s, %s]", vs, ls, us)
		} else {
			if min != nil {
				return fmt.Errorf("Value %s out of bounds [%s, +∞)", vs, ls)
			} else {
				return fmt.Errorf("Value %s out of bounds (-∞, %s]", vs, us)
			}
		}
	}
	return nil
}

// Checks whether d's length is within limits specified in the metadata.
// This is only applicable to strings.
func checkLength(d interface{}, t metadata) error {
	var obl, obh bool
	if t.constraints.minlen >= 0 || t.constraints.maxlen >= 0 {
		rv := reflect.ValueOf(d)
		vk := rv.Type().Kind()
		if vk == reflect.Ptr {
			rv = rv.Elem()
			vk = rv.Type().Kind()
		}
		vl := -1
		if vk == reflect.String { vl = rv.Len(); }
		if vl >= 0 {
			obl = t.constraints.minlen >= 0 && vl < t.constraints.minlen
			obh = t.constraints.maxlen >= 0 && vl > t.constraints.maxlen
		}
	}
	switch {
	case obl:
		return fmt.Errorf("Value is too short")
	case obh:
		return fmt.Errorf("Value is too long")
	}
	return nil
}

func checkConstraints(d interface{}, t metadata, bg boundaryGetter) error {
	if err := checkBounds(d, t, bg); err != nil { return err; }
	if err := checkLength(d, t); err != nil { return err; }
	return nil
}
