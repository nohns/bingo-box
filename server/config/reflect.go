package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type valueProviderFunc func(confField string, field reflect.StructField) (valuer, error)

type valuer interface {
	Value() (interface{}, error)
}

type provider interface {
	Provide(confField string, field reflect.StructField) (valuer, error)
}

type AfterProvider interface {
	AfterProvide() error
}

type reflectorField struct {
	confField string
	fv        *reflect.Value
	valr      valuer
}

type reflector struct {
	namespaces map[string]string
	fields     []reflectorField
	ps         []provider
}

func reflectVals(conf *Conf, providers ...provider) (*reflector, error) {
	v := reflect.Indirect(reflect.ValueOf(conf))
	if v.Type().Kind() != reflect.Struct {
		return nil, errors.New("conf: could not perform reflection on non-struct conf")
	}
	refl := &reflector{
		namespaces: make(map[string]string),
		fields:     make([]reflectorField, 0),
		ps:         providers,
	}
	err := refl.traverse("", v.Type().Name(), &v)
	if err != nil {
		return nil, err
	}

	// Run after provide hook
	for _, p := range refl.ps {
		afterP, ok := p.(AfterProvider)
		if ok {
			err := afterP.AfterProvide()
			if err != nil {
				return nil, err
			}
		}
	}

	// Set values of provided valuers
	for _, rf := range refl.fields {
		val, err := rf.valr.Value()
		if err != nil {
			return nil, err
		}
		if val == nil {
			continue
		}

		valt := reflect.TypeOf(val)
		if rf.fv.Kind() != valt.Kind() {
			return nil, fmt.Errorf("conf: value provided for '%s' was %v instead of %v", rf.confField, valt.Kind(), rf.fv.Kind())
		}

		rf.fv.Set(reflect.ValueOf(val))
	}

	return refl, nil
}

func (r *reflector) traverse(confPrefix string, ns string, v *reflect.Value) error {
	for i := 0; i < v.NumField(); i++ {
		ft := v.Type().Field(i)
		fv := v.Field(i)

		confField, ok := ft.Tag.Lookup("conf")
		if !ok {
			confField = confFieldFromName(ft.Name)
		}
		fullConfField := strings.TrimSpace(confPrefix + " " + confField)

		// set field in namespace
		nsField := ns + "." + ft.Name
		r.namespaces[nsField] = fullConfField

		switch fv.Kind() {
		case reflect.Struct:
			r.traverse(fullConfField, nsField, &fv)
		case reflect.String, reflect.Int, reflect.Bool:
			for _, p := range r.ps {
				valr, err := p.Provide(fullConfField, ft)
				if err != nil {
					return err
				}
				r.fields = append(r.fields, reflectorField{
					fv:   &fv,
					valr: valr,
				})
			}

		default:
			return fmt.Errorf("conf: unsupported type kind, %v, found in conf struct type", fv.Kind())
		}
	}

	return nil
}

func confFieldFromName(fname string) string {
	var confField string
	for i, c := range fname {
		lowerc := strings.ToLower(string(c))
		if i != 0 && string(c) != lowerc {
			confField += " "
		}

		confField += lowerc
	}

	return confField
}

type stringValuer struct {
	kind  reflect.Kind
	input string
}

func (sv *stringValuer) Value() (interface{}, error) {
	if val := sv.input; val != "" {
		switch sv.kind {
		case reflect.String:
			return val, nil
		case reflect.Bool:
			b := strings.ToLower(val)
			if b == "false" {
				return false, nil
			} else if b == "true" {
				return true, nil
			}
		case reflect.Int:
			i, err := strconv.Atoi(val)
			if err != nil {
				return nil, err
			}

			return i, nil
		default:
			return nil, fmt.Errorf("conf: type kind %s not supported by string valuer", sv.kind.String())
		}
	}

	return nil, nil
}
