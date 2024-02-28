package functions_test

import (
	"fmt"
	"testing"

	"github.com/playmixer/pc.assistent/functions"
)

func TestIsInt(t *testing.T) {
	cases := map[string]bool{
		"09": true,
	}

	for k, v := range cases {
		if functions.IsInt(k) != v {
			fmt.Printf("TEST %s is number ? %v != %v\n", k, v, functions.IsInt(k))
			t.Fail()
		}
	}
}

func TestStrToInt(t *testing.T) {
	cases := map[string]int{
		"09": 9,
	}
	for k, v := range cases {
		if functions.StrToInt(k) != v {
			fmt.Printf("TEST %s, %v not equale %v\n", k, v, functions.StrToInt(k))
			t.Fail()
		}
	}
}
