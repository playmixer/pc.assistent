package voskclient

import (
	"fmt"
	"testing"
)

type log struct {
}

func (l *log) INFO(v ...string) {
	fmt.Println(v)
}

func (l *log) ERROR(v ...string) {
	fmt.Println(v)
}

func (l *log) DEBUG(v ...string) {
	fmt.Println(v)
}

func TestPostConfigure(t *testing.T) {
	c := New(&log{})
	err := c.PostConfigure()
	if err != nil {
		t.Fatal(err)
	}
}
