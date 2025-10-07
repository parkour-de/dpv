package security

import "testing"

func TestIsStrongPassword(t *testing.T) {
	cases := []struct {
		password string
		valid    bool
		errMsg   string
	}{
		{"1234567890", false, "must not be only digits"},
		{"abcdefghij", false, "must not be only lowercase letters"},
		{"ABCDEFGHIJ", false, "must not be only uppercase letters"},
		{"abcABC123!@", true, ""},
		{"aA1!aA1!aA1!", false, "must have at least 8 different glyphs"},
		{"aA1!aA1!", false, "too short (min 10 characters), must have at least 8 different glyphs"},
		{"short1A!@", false, "too short (min 10 characters)"},
		{"abcdefghijklmno", false, "must not be only lowercase letters"},
		{"ABCDEFGHIJKLMNO", false, "must not be only uppercase letters"},
		{"123456789012345", false, "must not be only digits"},
		{"abcABC123!@#", true, ""},
	}
	for _, c := range cases {
		ok, err := IsStrongPassword(c.password)
		if ok != c.valid {
			t.Errorf("IsStrongPassword(%q) = %v, want %v", c.password, ok, c.valid)
		}
		if !ok && err != nil && c.errMsg != "" && err.Error() != c.errMsg {
			t.Errorf("IsStrongPassword(%q) error = %q, want %q", c.password, err.Error(), c.errMsg)
		}
	}
}
