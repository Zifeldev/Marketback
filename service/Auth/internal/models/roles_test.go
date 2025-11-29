package models

import "testing"

func TestValidateRole(t *testing.T) {
	cases := []struct {
		role  string
		valid bool
	}{
		{"user", true},
		{"seller", true},
		{"admin", true},
		{"", false},
		{"unknown", false},
	}

	for _, c := range cases {
		if (ValidateRole(c.role) == nil) != c.valid {
			t.Errorf("role %q validity expected %v", c.role, c.valid)
		}
	}
}

func TestIsValidRole(t *testing.T) {
	if !IsValidRole(RoleUser) || !IsValidRole(RoleSeller) || !IsValidRole(RoleAdmin) {
		t.Fatalf("expected core roles to be valid")
	}
	if IsValidRole("") || IsValidRole("nope") {
		t.Fatalf("expected invalid roles to be invalid")
	}
}
