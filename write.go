package gcfg

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

func writeItem(v reflect.Value, name string, w io.Writer) error {
	vi := v.Interface()
	z := reflect.Zero(v.Type())
	if !reflect.DeepEqual(z.Interface(), vi) {
		if p, ok := typeFormatters[v.Type()]; ok {
			vi = p(vi)
		} else {
			if _, ok := vi.(fmt.Stringer); !ok && v.CanAddr() {
				vr := v.Addr()
				vri := vr.Interface()
				if _, okk := vri.(fmt.Stringer); okk {
					vi = vri
				}
			}
		}
		if _, err := w.Write([]byte(fmt.Sprintf("%s = %v\n", name, vi))); err != nil {
			return err
		}
	}
	return nil
}

func writeInSection(vSect reflect.Value, w io.Writer) error {
	tp := vSect.Type()
	if tp.Kind() == reflect.Ptr {
		if vSect.IsNil() {
			return nil
		}
		vSect = vSect.Elem()
		tp = vSect.Type()
	}
	for i, n := 0, vSect.NumField(); i < n; i++ {
		vVar := vSect.Field(i)
		if !vVar.IsValid() {
			continue
		}
		sf := tp.Field(i)
		if sf.Anonymous {
			if err := writeInSection(vVar, w); err != nil {
				return err
			}
			continue
		}
		t := newMetadata(sf.Tag.Get("gcfg"), sf.Tag)
		in := t.ident
		if in == "-" {
			continue
		}
		if in == "" {
			in = strings.ToLower(sf.Name)
		}
		isMulti := vVar.Type().Name() == "" && vVar.Kind() == reflect.Slice
		if !isMulti {
			writeItem(vVar, in, w)
		} else {
			for i, n := 0, vVar.Len(); i < n; i++ {
				writeItem(vVar.Index(i), in, w)
			}
		}
	}
	return nil
}

func writeSection(vSect reflect.Value, name string, w io.Writer) error {
	if _, err := w.Write([]byte("[")); err != nil {
		return err
	}
	if _, err := w.Write([]byte(name)); err != nil {
		return err
	}
	if _, err := w.Write([]byte("]\n")); err != nil {
		return err
	}
	if err := writeInSection(vSect, w); err != nil {
		return err
	}
	if _, err := w.Write([]byte("\n")); err != nil {
		return err
	}
	return nil
}

func write(vc reflect.Value, w io.Writer) error {
	vt := vc.Type()
	for i, n := 0, vc.NumField(); i < n; i++ {
		vSect := vc.Field(i)
		if !vSect.IsValid() {
			continue
		}
		isMap := false
		if isMap = vSect.Kind() == reflect.Map; isMap {
			vst := vSect.Type()
			if vst.Key().Kind() != reflect.String ||
				vst.Elem().Kind() != reflect.Ptr ||
				vst.Elem().Elem().Kind() != reflect.Struct {
				continue
			}
			if vSect.IsNil() {
				continue
			}
		} else if vSect.Kind() == reflect.Ptr && (vSect.IsNil() || vSect.Elem().Kind() == reflect.Struct) {
			if vSect.IsNil() {
				continue
			}
			vSect = vSect.Elem()
		} else if vSect.Kind() != reflect.Struct {
			continue
		}
		sf := vt.Field(i)
		if sf.Anonymous {
			if err := write(vSect, w); err != nil {
				return err
			}
			continue
		}
		t := newMetadata(sf.Tag.Get("gcfg"), sf.Tag)
		name := t.ident
		if name == "" {
			name = strings.ToLower(sf.Name)
		}
		if isMap {
			for _, k := range vSect.MapKeys() {
				n := fmt.Sprintf("%s \"%s\"", name, k)
				if err := writeSection(vSect.MapIndex(k).Elem(), n, w); err != nil {
					return err
				}
			}
		} else {
			if err := writeSection(vSect, name, w); err != nil {
				return err
			}
		}
	}
	return nil
}

// Write writes config in gcfg formatted data.
func Write(config interface{}, w io.Writer) error {
	vpc := reflect.ValueOf(config)
	if vpc.Kind() != reflect.Ptr || vpc.Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf("config must be a pointer to a struct"))
	}
	vc := vpc.Elem()
	return write(vc, w)
}

type TypeFormatter func(interface{}) string

var typeFormatters = map[reflect.Type]TypeFormatter{}

func RegisterTypeFormatter(tgtType reflect.Type, typeFormatter TypeFormatter) error {
	typeFormatters[tgtType] = typeFormatter
	return nil
}
