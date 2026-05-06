package main

import "testing"

func TestIsVersionRequest(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want bool
	}{
		{name: "double dash", args: []string{"--version"}, want: true},
		{name: "single dash", args: []string{"-version"}, want: true},
		{name: "plain version", args: []string{"version"}, want: true},
		{name: "query content", args: []string{"hello world"}, want: false},
		{name: "multiple args", args: []string{"version", "extra"}, want: false},
	}
	for _, tc := range cases {
		if got := isVersionRequest(tc.args); got != tc.want {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.want, got)
		}
	}
}
