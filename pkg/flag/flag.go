package flag

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Allocate a flagset, bind it to val and return the flag set.
func GetFlagSet(name string, val interface{}) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.PanicOnError)
	RegisterFlags(name, val, fs)
	fs.Usage = func() {
		fs.PrintDefaults()
	}
	return fs
}

// Register fields in the given struct that have the tag `flag:"name,desc"`.
// Nested structs are supported as long as the field is a struct value field and not pointer to a struct.
// Exception to this is the use of StringList which needs to be a pointer.  The StringList type implements
// the Set and String methods required by the flag package and is dynamically allocated when registering its flag.
// See the test case for example.
func RegisterFlags(name string, val interface{}, fs *flag.FlagSet) {
	t := reflect.TypeOf(val).Elem()
	v := reflect.Indirect(reflect.ValueOf(val)) // the actual value of val
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			RegisterFlags(name+"."+field.Name, v.Field(i).Addr().Interface(), fs)
			continue
		}

		// See https://golang.org/ref/spec#Uniqueness_of_identifiers
		exported := field.PkgPath == ""
		if exported {

			tag := field.Tag
			spec := tag.Get("flag")
			if spec == "" {
				continue
			}

			// Bind the flag based on the tag spec
			f, d := "", ""
			p := strings.Split(spec, ",")
			if len(p) == 1 {
				// Just one field, use it as description
				f = fmt.Sprintf("%s.%s", name, strings.ToLower(field.Name))
				d = strings.Trim(p[0], " ")
			} else {
				// More than one, the first is the name of the flag
				f = strings.Trim(p[0], " ")
				d = strings.Trim(p[1], " ")
			}

			fv := v.Field(i).Interface()
			if v.Field(i).CanAddr() {
				ptr := v.Field(i).Addr().Interface() // The pointer value

				switch fv := fv.(type) {
				case bool:
					fs.BoolVar(ptr.(*bool), f, fv, d)
				case []bool:
					if len(fv) == 0 {
						// Special case where we allocate an empty list - otherwise it's default.
						v.Field(i).Set(reflect.ValueOf([]bool{}))
					}
					fs.Var(&boolListProxy{list: ptr.(*[]bool)}, f, d)
				case time.Duration:
					fs.DurationVar(ptr.(*time.Duration), f, fv, d)
				case []time.Duration:
					if len(fv) == 0 {
						// Special case where we allocate an empty list - otherwise it's default.
						v.Field(i).Set(reflect.ValueOf([]time.Duration{}))
					}
					fs.Var(&durationListProxy{list: ptr.(*[]time.Duration)}, f, d)
				case float64:
					fs.Float64Var(ptr.(*float64), f, fv, d)
				case []float64:
					if len(fv) == 0 {
						// Special case where we allocate an empty list - otherwise it's default.
						v.Field(i).Set(reflect.ValueOf([]float64{}))
					}
					fs.Var(&float64ListProxy{list: ptr.(*[]float64)}, f, d)
				case int:
					fs.IntVar(ptr.(*int), f, fv, d)
				case []int:
					if len(fv) == 0 {
						// Special case where we allocate an empty list - otherwise it's default.
						v.Field(i).Set(reflect.ValueOf([]int{}))
					}
					fs.Var(&intListProxy{list: ptr.(*[]int)}, f, d)
				case int64:
					fs.Int64Var(ptr.(*int64), f, fv, d)
				case []int64:
					if len(fv) == 0 {
						// Special case where we allocate an empty list - otherwise it's default.
						v.Field(i).Set(reflect.ValueOf([]int64{}))
					}
					fs.Var(&int64ListProxy{list: ptr.(*[]int64)}, f, d)
				case string:
					fs.StringVar(ptr.(*string), f, fv, d)
				case []string:
					if len(fv) == 0 {
						// Special case where we allocate an empty list - otherwise it's default.
						v.Field(i).Set(reflect.ValueOf([]string{}))
					}
					fs.Var(&stringListProxy{list: ptr.(*[]string)}, f, d)
				case uint:
					fs.UintVar(ptr.(*uint), f, fv, d)
				case []uint:
					if len(fv) == 0 {
						// Special case where we allocate an empty list - otherwise it's default.
						v.Field(i).Set(reflect.ValueOf([]uint{}))
					}
					fs.Var(&uintListProxy{list: ptr.(*[]uint)}, f, d)
				case uint64:
					fs.Uint64Var(ptr.(*uint64), f, fv, d)
				case []uint64:
					if len(fv) == 0 {
						// Special case where we allocate an empty list - otherwise it's default.
						v.Field(i).Set(reflect.ValueOf([]uint64{}))
					}
					fs.Var(&uint64ListProxy{list: ptr.(*[]uint64)}, f, d)
				default:
					// We only register if the field is a concrete vale and not a pointer
					// since we don't automatically allocate zero value structs to fill the field slot.
					switch field.Type.Kind() {
					case reflect.Struct:
						RegisterFlags(f, ptr, fs)
					case reflect.Slice:
						// TODO - should refactor to use the generic sliceProxy instead of the typed slice proxies above.
						et := field.Type.Elem()
						proxy := &sliceProxy{
							fieldType: field.Type,
							elemType:  et,
							slice:     ptr,
							defaults:  reflect.ValueOf(fv).Len() > 0,
							toString: func(v interface{}) string {
								return fmt.Sprint("%v", v)
							},
						}
						fs.Var(proxy, f, d)
						switch {
						// Checking for string is placed here first because other types are
						// convertible to string as well.
						case reflect.TypeOf(string("")).ConvertibleTo(et):
							proxy.fromString = func(s string) (interface{}, error) {
								return s, nil
							}
							proxy.toString = func(v interface{}) string {
								return v.(string)
							}
						case reflect.TypeOf(bool(true)).ConvertibleTo(et):
							proxy.fromString = func(s string) (interface{}, error) {
								value, err := strconv.ParseBool(s)
								if err != nil {
									return false, err
								}
								return value, nil
							}
						case reflect.TypeOf(float64(1.)).ConvertibleTo(et):
							proxy.fromString = func(s string) (interface{}, error) {
								value, err := strconv.ParseFloat(s, 64)
								if err != nil {
									return float64(0), err
								}
								return value, nil
							}
							proxy.toString = func(v interface{}) string {
								return v.(string)
							}
						case reflect.TypeOf(int(1)).ConvertibleTo(et):
							proxy.fromString = func(s string) (interface{}, error) {
								value, err := strconv.Atoi(s)
								if err != nil {
									return int(0), err
								}
								return value, nil
							}
						case reflect.TypeOf(int64(1)).ConvertibleTo(et):
							proxy.fromString = func(s string) (interface{}, error) {
								value, err := strconv.ParseInt(s, 10, 64)
								if err != nil {
									return int64(0), err
								}
								return value, nil
							}
							proxy.toString = func(v interface{}) string {
								return v.(string)
							}
						case reflect.TypeOf(uint(1)).ConvertibleTo(et):
							proxy.fromString = func(s string) (interface{}, error) {
								value, err := strconv.ParseUint(s, 10, 32)
								if err != nil {
									return uint(0), err
								}
								return value, nil
							}
						case reflect.TypeOf(uint64(1)).ConvertibleTo(et):
							proxy.fromString = func(s string) (interface{}, error) {
								value, err := strconv.ParseUint(s, 10, 64)
								if err != nil {
									return uint64(0), err
								}
								return value, nil
							}
						case reflect.TypeOf(time.Second).ConvertibleTo(et):
							proxy.fromString = func(s string) (interface{}, error) {
								value, err := time.ParseDuration(s)
								if err != nil {
									return false, err
								}
								return value, nil
							}
						}
					}
				}
			}
		}
	}
}

var (
	stringType = reflect.TypeOf("")
)

// For a list of types that are convertible to string
type sliceProxy struct {
	fieldType  reflect.Type
	elemType   reflect.Type                      // the element type
	fromString func(string) (interface{}, error) // conversion from string
	toString   func(interface{}) string          // to string
	slice      interface{}                       // the Pointer to the slice
	defaults   bool                              // set to true on first time Set is called.
}

func (this *sliceProxy) Set(value string) error {
	v, err := this.fromString(value)
	if err != nil {
		return err
	}
	newElement := reflect.ValueOf(reflect.ValueOf(v).Convert(this.elemType).Interface())
	if this.defaults {
		reflect.ValueOf(this.slice).Elem().Set(reflect.Zero(this.fieldType))
		this.defaults = false
	}
	reflect.ValueOf(this.slice).Elem().Set(reflect.Append(reflect.ValueOf(this.slice).Elem(), newElement))
	return nil
}
func (this *sliceProxy) String() string {
	list := []string{}
	for i := 0; i < reflect.ValueOf(this.slice).Elem().Len(); i++ {
		str := this.toString(reflect.ValueOf(this.slice).Elem().Index(i).Interface())
		list = append(list, str)
		// ev := reflect.ValueOf(this.slice).Elem().Index(i).Convert(stringType).Interface()
		// list = append(list, ev.(string))
	}
	return strings.Join(list, ",")
}

// Supports default values.  This means that if the slice was initialized with value, setting
// via flag will wipe out the existing value.
type stringListProxy struct {
	list *[]string
	set  bool // set to true on first time Set is called.
}

func (this *stringListProxy) Set(value string) error {
	if this.set {
		*this.list = append(*this.list, value)
	} else {
		// false means we have default value, now wipe it out
		*this.list = []string{value}
		this.set = true
	}
	return nil
}
func (this *stringListProxy) String() string {
	return strings.Join(*this.list, ",")
}

// Supports default values.  This means that if the slice was initialized with value, setting
// via flag will wipe out the existing value.
type intListProxy struct {
	list *[]int
	set  bool // set to true on first time Set is called.
}

func (this *intListProxy) Set(str string) error {
	value, err := strconv.Atoi(str)
	if err != nil {
		return err
	}
	if this.set {
		*this.list = append(*this.list, value)
	} else {
		// false means we have default value, now wipe it out
		*this.list = []int{value}
		this.set = true
	}
	return nil
}
func (this *intListProxy) String() string {
	list := make([]string, len(*this.list))
	for i, v := range *this.list {
		list[i] = strconv.Itoa(v)
	}
	return strings.Join(list, ",")
}

// Supports default values.  This means that if the slice was initialized with value, setting
// via flag will wipe out the existing value.
type int64ListProxy struct {
	list *[]int64
	set  bool // set to true on first time Set is called.
}

func (this *int64ListProxy) Set(str string) error {
	value, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return err
	}
	if this.set {
		*this.list = append(*this.list, value)
	} else {
		// false means we have default value, now wipe it out
		*this.list = []int64{value}
		this.set = true
	}
	return nil
}
func (this *int64ListProxy) String() string {
	list := make([]string, len(*this.list))
	for i, v := range *this.list {
		list[i] = strconv.FormatInt(v, 10)
	}
	return strings.Join(list, ",")
}

// Supports default values.  This means that if the slice was initialized with value, setting
// via flag will wipe out the existing value.
type float64ListProxy struct {
	list *[]float64
	set  bool // set to true on first time Set is called.
}

func (this *float64ListProxy) Set(str string) error {
	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return err
	}
	if this.set {
		*this.list = append(*this.list, value)
	} else {
		// false means we have default value, now wipe it out
		*this.list = []float64{value}
		this.set = true
	}
	return nil
}
func (this *float64ListProxy) String() string {
	list := make([]string, len(*this.list))
	for i, v := range *this.list {
		list[i] = strconv.FormatFloat(v, 'E', -1, 64)
	}
	return strings.Join(list, ",")
}

// Supports default values.  This means that if the slice was initialized with value, setting
// via flag will wipe out the existing value.
type boolListProxy struct {
	list *[]bool
	set  bool // set to true on first time Set is called.
}

func (this *boolListProxy) Set(str string) error {
	value, err := strconv.ParseBool(str)
	if err != nil {
		return err
	}
	if this.set {
		*this.list = append(*this.list, value)
	} else {
		// false means we have default value, now wipe it out
		*this.list = []bool{value}
		this.set = true
	}
	return nil
}
func (this *boolListProxy) String() string {
	list := make([]string, len(*this.list))
	for i, v := range *this.list {
		list[i] = strconv.FormatBool(v)
	}
	return strings.Join(list, ",")
}

// Supports default values.  This means that if the slice was initialized with value, setting
// via flag will wipe out the existing value.
type uint64ListProxy struct {
	list *[]uint64
	set  bool // set to true on first time Set is called.
}

func (this *uint64ListProxy) Set(str string) error {
	value, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return err
	}
	if this.set {
		*this.list = append(*this.list, value)
	} else {
		// false means we have default value, now wipe it out
		*this.list = []uint64{value}
		this.set = true
	}
	return nil
}
func (this *uint64ListProxy) String() string {
	list := make([]string, len(*this.list))
	for i, v := range *this.list {
		list[i] = strconv.FormatUint(v, 10)
	}
	return strings.Join(list, ",")
}

// Supports default values.  This means that if the slice was initialized with value, setting
// via flag will wipe out the existing value.
type uintListProxy struct {
	list *[]uint
	set  bool // set to true on first time Set is called.
}

func (this *uintListProxy) Set(str string) error {
	value, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return err
	}
	if this.set {
		*this.list = append(*this.list, uint(value))
	} else {
		// false means we have default value, now wipe it out
		*this.list = []uint{uint(value)}
		this.set = true
	}
	return nil
}
func (this *uintListProxy) String() string {
	list := make([]string, len(*this.list))
	for i, v := range *this.list {
		list[i] = strconv.FormatUint(uint64(v), 10)
	}
	return strings.Join(list, ",")
}

// Supports default values.  This means that if the slice was initialized with value, setting
// via flag will wipe out the existing value.
type durationListProxy struct {
	list *[]time.Duration
	set  bool // set to true on first time Set is called.
}

func (this *durationListProxy) Set(str string) error {
	value, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	if this.set {
		*this.list = append(*this.list, value)
	} else {
		// false means we have default value, now wipe it out
		*this.list = []time.Duration{value}
		this.set = true
	}
	return nil
}
func (this *durationListProxy) String() string {
	list := make([]string, len(*this.list))
	for i, v := range *this.list {
		list[i] = v.String()
	}
	return strings.Join(list, ",")
}
