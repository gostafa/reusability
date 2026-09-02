package cli

import (
	"reflect"
	"testing"
)

// White-box: comma-list parsing trims blanks and empties.
func TestSplitList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"a,b,c", []string{"a", "b", "c"}},
		{" a , ,b ", []string{"a", "b"}},
	}
	for _, tt := range tests {
		if got := splitList(tt.in); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("splitList(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

// White-box: the raw-args scan behind --help --web accepts truthy -web
// tokens before a "--" terminator and nothing else.
func TestWantsWebHelp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args []string
		want bool
	}{
		{[]string{"--help", "--web"}, true},
		{[]string{"--web", "--help"}, true},
		{[]string{"-web", "-h"}, true},
		{[]string{"--web=true", "--help"}, true},
		{[]string{"--web=false", "--help"}, false},
		{[]string{"--web=nonsense", "--help"}, false},
		{[]string{"--help"}, false},
		{[]string{"--help", "--", "--web"}, false},
		{nil, false},
	}
	for _, tt := range tests {
		if got := wantsWebHelp(tt.args); got != tt.want {
			t.Errorf("wantsWebHelp(%v) = %v, want %v", tt.args, got, tt.want)
		}
	}
}

// White-box: bad flags and formats exit with code 2 before any loading.
func TestRunUsageErrors(t *testing.T) {
	t.Parallel()

	if code := run([]string{"--format=xml"}); code != 2 {
		t.Errorf("invalid format exit = %d, want 2", code)
	}

	if code := run([]string{"--this-flag-does-not-exist"}); code != 2 {
		t.Errorf("unknown flag exit = %d, want 2", code)
	}
}
