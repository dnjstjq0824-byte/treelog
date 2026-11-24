package main

import (
	"reflect"
	"testing"
)

func TestPublicHosts(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"https://blog.example.com", []string{"blog.example.com"}},
		{"http://localhost:3000", []string{"localhost"}},
		{"blog.example.com", []string{"blog.example.com"}},
		{"", nil},
		{"http://", nil},
	}
	for _, tc := range cases {
		if got := publicHosts(tc.in); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("publicHosts(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
