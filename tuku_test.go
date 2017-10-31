package main

import "testing"

var tests = []struct {
	A interface{}
	B interface{}
}{
	{"/beep/baap/boop.txt", "boop.txt"},
	{"fing/fan/foo", "foo"},
}

func TestPluck(t *testing.T) {
	for _, test := range tests {
		if fn := pluck(test.A.(string), "/"); fn != test.B {
			t.Fatalf("want: %s, got: %s\n", test.B, fn)
		}
	}
}
