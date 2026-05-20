package update

import "testing"

func TestCompare(t *testing.T) {
	t.Parallel()

	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"0.1.1", "0.1.0", 1},
		{"2.0.0", "1.9.9", 1},
	}
	for _, tc := range cases {
		got := Compare(tc.a, tc.b)
		if got != tc.want {
			t.Fatalf("Compare(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestIsNewer(t *testing.T) {
	t.Parallel()

	if !IsNewer("0.1.1", "0.1.0") {
		t.Fatal("expected 0.1.1 to be newer than 0.1.0")
	}
	if IsNewer("0.1.0", "0.1.1") {
		t.Fatal("expected 0.1.0 not to be newer than 0.1.1")
	}
	if !IsNewer("0.2.0", "dev") {
		t.Fatal("expected tagged release to be newer than dev")
	}
}
