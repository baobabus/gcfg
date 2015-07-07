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

type cRegTypes struct {
	Reg_Types_1   cRegTypes1
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
	for _, tt := range []struct {
		gcfg string
		exp  interface{}
		ok   bool
	}{
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
}
