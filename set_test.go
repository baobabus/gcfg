package gcfg

import (
	"testing"
	"time"
	"reflect"
	"net/mail"
)

type cRegTypes1 struct {
	Duration1   time.Duration
	Duration2 []time.Duration
	Email1     *mail.Address 
	Email2      mail.Address 
	Email3   []*mail.Address 
	Email4    []mail.Address 
}

type cBoundsTypes1 struct {
	IntR1       int           `min:"10" max:"20"`
	IntR2     []int           `min:"10" max:"20"`
	IntL1       int           `min:"10"`
	IntU1       int           `max:"20"`
	FloatR1     float32       `min:"10.0" max:"20.0"`
	StringR1    string        `min:"b" max:"zz"`
	TimeR1      time.Time     `min:"2000-01-01T00:00:00.00Z" max:"2000-01-02T00:00:00.00Z"`
	DurationR1  time.Duration `min:"1h" max:"1h30m"`
}

type cRegTypes struct {
	Reg_Types_1    cRegTypes1
	Bounds_Types_1 cBoundsTypes1
}

type stTestCase struct {
	gcfg string
	exp  interface{}
	ok   bool
}

func assert(tt *stTestCase, t *testing.T) {
	res := &cRegTypes{}
	err := ReadStringInto(res, tt.gcfg)
	if tt.ok {
		if err != nil {
			t.Errorf("%s fail: got error %v, wanted ok", tt.gcfg, err)
			return
		} else if !reflect.DeepEqual(res, tt.exp) {
			t.Errorf("%s fail: got value %#v, wanted value %#v", tt.gcfg, res, tt.exp)
			return
		}
		if !testing.Short() {
			t.Logf("%s pass: got value %#v", tt.gcfg, res)
		}
	} else { // !tt.ok
		if err == nil {
			t.Errorf("%s fail: got value %#v, wanted error", tt.gcfg, res)
			return
		}
		if !testing.Short() {
			t.Logf("%s pass: got error %v", tt.gcfg, err)
		}
	}
}

func TestMissignTypeParser(t *testing.T) {
	for _, tt := range []stTestCase{
		{"[reg-types-1]\nduration1=5m", &cRegTypes{Reg_Types_1: cRegTypes1{}}, false},
		{"[reg-types-1]\nduration1=5", &cRegTypes{Reg_Types_1: cRegTypes1{Duration1: time.Duration(5)}}, true},
		{"[reg-types-1]\nemail1=foo@bar.com", &cRegTypes{Reg_Types_1: cRegTypes1{}}, false},
	} {
		assert(&tt, t)
	}
}

func TestRegisteredTypeParser(t *testing.T) {
	var d time.Duration;
	RegisterTypeParser(reflect.TypeOf(d), func(blank bool, val string) (interface{}, error) {
		if blank {
			return nil, nil
		}
		return time.ParseDuration(val)
	})
	RegisterTypeParser(reflect.TypeOf(mail.Address{}), func(blank bool, val string) (interface{}, error) {
		if blank {
			return nil, nil
		}
		return mail.ParseAddress(val)
	})
	for _, tt := range []stTestCase{
		{"[reg-types-1]\nduration1=5m", &cRegTypes{Reg_Types_1: cRegTypes1{Duration1: 5 * time.Minute}}, true},
		{"[reg-types-1]\nduration1=1m30s", &cRegTypes{Reg_Types_1: cRegTypes1{Duration1: 90 * time.Second}}, true},
		{"[reg-types-1]\nduration1=1m1m", &cRegTypes{Reg_Types_1: cRegTypes1{Duration1: time.Duration(2 * time.Minute)}}, true},
		{"[reg-types-1]\nduration1=30", &cRegTypes{Reg_Types_1: cRegTypes1{Duration1: time.Duration(0)}}, false},
		{"[reg-types-1]\nduration1=m", &cRegTypes{Reg_Types_1: cRegTypes1{Duration1: time.Duration(0)}}, false},
		{"[reg-types-1]\nduration2=5m", &cRegTypes{Reg_Types_1: cRegTypes1{Duration2: []time.Duration{5 * time.Minute}}}, true},
		{"[reg-types-1]\nemail1=foo@bar.com", &cRegTypes{Reg_Types_1: cRegTypes1{Email1: &mail.Address{"", "foo@bar.com"}}}, true},
		{"[reg-types-1]\nemail1=<foo@bar.com>", &cRegTypes{Reg_Types_1: cRegTypes1{Email1: &mail.Address{"", "foo@bar.com"}}}, true},
		{"[reg-types-1]\nemail1=\"foo bar\" <foo@bar.com>", &cRegTypes{Reg_Types_1: cRegTypes1{Email1: &mail.Address{"foo bar", "foo@bar.com"}}}, true},
		{"[reg-types-1]\nemail1=foo bar <foo@bar.com>", &cRegTypes{Reg_Types_1: cRegTypes1{Email1: &mail.Address{"foo bar", "foo@bar.com"}}}, true},
		{"[reg-types-1]\nemail1=foo bar  <foo@bar.com>", &cRegTypes{Reg_Types_1: cRegTypes1{Email1: &mail.Address{"foo bar", "foo@bar.com"}}}, true},
		{"[reg-types-1]\nemail1=<foo bar> <foo@bar.com>", &cRegTypes{Reg_Types_1: cRegTypes1{}}, false},
		{"[reg-types-1]\nemail1=<foo@foo@bar.com>", &cRegTypes{Reg_Types_1: cRegTypes1{}}, false},
		{"[reg-types-1]\nemail2=foo@bar.com", &cRegTypes{Reg_Types_1: cRegTypes1{Email2: mail.Address{"", "foo@bar.com"}}}, true},
		{"[reg-types-1]\nemail3=foo@bar.com", &cRegTypes{Reg_Types_1: cRegTypes1{Email3: []*mail.Address{&mail.Address{"", "foo@bar.com"}}}}, true},
		{"[reg-types-1]\nemail4=foo@bar.com", &cRegTypes{Reg_Types_1: cRegTypes1{Email4: []mail.Address{mail.Address{"", "foo@bar.com"}}}}, true},
	} {
		assert(&tt, t)
	}
}

func TestBoundsConstraints(t *testing.T) {
	var d time.Duration;
	RegisterTypeParser(reflect.TypeOf(d), func(blank bool, val string) (interface{}, error) {
		if blank {
			return nil, nil
		}
		return time.ParseDuration(val)
	})
	for _, tt := range []stTestCase{
		{"[bounds-types-1]\nintR1=10", &cRegTypes{Bounds_Types_1: cBoundsTypes1{IntR1: 10}}, true},
		{"[bounds-types-1]\nintR1=15", &cRegTypes{Bounds_Types_1: cBoundsTypes1{IntR1: 15}}, true},
		{"[bounds-types-1]\nintR1=20", &cRegTypes{Bounds_Types_1: cBoundsTypes1{IntR1: 20}}, true},
		{"[bounds-types-1]\nintR1=21", &cRegTypes{Bounds_Types_1: cBoundsTypes1{IntR1: 21}}, false},
		{"[bounds-types-1]\nintR1=9",  &cRegTypes{Bounds_Types_1: cBoundsTypes1{IntR1: 9}}, false},

		{"[bounds-types-1]\nintR2=10\nintR2=20", &cRegTypes{Bounds_Types_1: cBoundsTypes1{IntR2: []int{10, 20}}}, true},
		{"[bounds-types-1]\nintR2=10\nintR2=21", &cRegTypes{Bounds_Types_1: cBoundsTypes1{IntR2: []int{10, 21}}}, false},

		{"[bounds-types-1]\nintL1=10", &cRegTypes{Bounds_Types_1: cBoundsTypes1{IntL1: 10}}, true},
		{"[bounds-types-1]\nintL1=15", &cRegTypes{Bounds_Types_1: cBoundsTypes1{IntL1: 15}}, true},
		{"[bounds-types-1]\nintL1=9",  &cRegTypes{Bounds_Types_1: cBoundsTypes1{IntL1: 9}}, false},
		{"[bounds-types-1]\nintU1=15", &cRegTypes{Bounds_Types_1: cBoundsTypes1{IntU1: 15}}, true},
		{"[bounds-types-1]\nintU1=20", &cRegTypes{Bounds_Types_1: cBoundsTypes1{IntU1: 20}}, true},
		{"[bounds-types-1]\nintU1=21",  &cRegTypes{Bounds_Types_1: cBoundsTypes1{IntU1: 21}}, false},

		{"[bounds-types-1]\nfloatR1=10", &cRegTypes{Bounds_Types_1: cBoundsTypes1{FloatR1: 10}}, true},
		{"[bounds-types-1]\nfloatR1=15", &cRegTypes{Bounds_Types_1: cBoundsTypes1{FloatR1: 15}}, true},
		{"[bounds-types-1]\nfloatR1=20", &cRegTypes{Bounds_Types_1: cBoundsTypes1{FloatR1: 20}}, true},
		{"[bounds-types-1]\nfloatR1=21", &cRegTypes{Bounds_Types_1: cBoundsTypes1{FloatR1: 21}}, false},
		{"[bounds-types-1]\nfloatR1=9",  &cRegTypes{Bounds_Types_1: cBoundsTypes1{FloatR1: 9}}, false},
		{"[bounds-types-1]\nstringR1=b", &cRegTypes{Bounds_Types_1: cBoundsTypes1{StringR1: "b"}}, true},
		{"[bounds-types-1]\nstringR1=dd", &cRegTypes{Bounds_Types_1: cBoundsTypes1{StringR1: "dd"}}, true},
		{"[bounds-types-1]\nstringR1=zz", &cRegTypes{Bounds_Types_1: cBoundsTypes1{StringR1: "zz"}}, true},
		{"[bounds-types-1]\nstringR1=a", &cRegTypes{Bounds_Types_1: cBoundsTypes1{StringR1: "a"}}, false},
		{"[bounds-types-1]\nstringR1=zza",  &cRegTypes{Bounds_Types_1: cBoundsTypes1{StringR1: "zza"}}, false},

		{"[bounds-types-1]\ntimeR1=2000-01-01T00:00:00.00Z", &cRegTypes{Bounds_Types_1: cBoundsTypes1{TimeR1: time.Date(2000,  1,  1, 00, 00, 00, 00, time.UTC)}}, true},
		{"[bounds-types-1]\ntimeR1=2000-01-01T10:00:00.00Z", &cRegTypes{Bounds_Types_1: cBoundsTypes1{TimeR1: time.Date(2000,  1,  1, 10, 00, 00, 00, time.UTC)}}, true},
		{"[bounds-types-1]\ntimeR1=2000-01-02T00:00:00.00Z", &cRegTypes{Bounds_Types_1: cBoundsTypes1{TimeR1: time.Date(2000,  1,  2, 00, 00, 00, 00, time.UTC)}}, true},
		{"[bounds-types-1]\ntimeR1=1999-12-31T00:00:00.00Z", &cRegTypes{Bounds_Types_1: cBoundsTypes1{TimeR1: time.Date(1999, 12, 31, 00, 00, 00, 00, time.UTC)}}, false},
		{"[bounds-types-1]\ntimeR1=2000-02-02T00:00:00.00Z", &cRegTypes{Bounds_Types_1: cBoundsTypes1{TimeR1: time.Date(2000,  2,  2, 00, 00, 00, 00, time.UTC)}}, false},

		{"[bounds-types-1]\ndurationR1=1h",    &cRegTypes{Bounds_Types_1: cBoundsTypes1{DurationR1: time.Hour                 }}, true},
		{"[bounds-types-1]\ndurationR1=1h10m", &cRegTypes{Bounds_Types_1: cBoundsTypes1{DurationR1: time.Hour + 10*time.Minute}}, true},
		{"[bounds-types-1]\ndurationR1=1h30m", &cRegTypes{Bounds_Types_1: cBoundsTypes1{DurationR1: time.Hour + 30*time.Minute}}, true},
		{"[bounds-types-1]\ndurationR1=55m",   &cRegTypes{Bounds_Types_1: cBoundsTypes1{DurationR1:             55*time.Minute}}, false},
		{"[bounds-types-1]\ndurationR1=1h35m", &cRegTypes{Bounds_Types_1: cBoundsTypes1{DurationR1: time.Hour + 35*time.Minute}}, false},
	} {
		assert(&tt, t)
	}
}
