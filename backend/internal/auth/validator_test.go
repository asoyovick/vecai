package auth

import "testing"

func TestValidateRegistration(t *testing.T) {
	if _, err := validateRegistration("David Ochieng", "david@vecai.co.ke", "demo1234"); err != nil {
		t.Fatalf("valid registration rejected: %v", err)
	}
	cases := []struct{ name, email, password, want string }{
		{"", "a@b.co", "passw0rd1", "name"},
		{"David", "not-an-email", "passw0rd1", "email"},
		{"David", "a@b.co", "short", "password"},
		{"David", "a@b.co", "onlyletters", "password"},
		{"David", "a@b.co", "12345678", "password"},
	}
	for _, c := range cases {
		_, err := validateRegistration(c.name, c.email, c.password)
		if err == nil {
			t.Fatalf("registration %+v accepted", c)
		}
	}
}

func TestValidateNewPassword(t *testing.T) {
	if err := ValidateNewPassword("newpass1"); err != nil {
		t.Fatalf("valid password rejected: %v", err)
	}
	for _, p := range []string{"", "short", "allletters", "12345678", string(make([]byte, 300))} {
		if err := ValidateNewPassword(p); err == nil {
			t.Fatalf("weak password %q accepted", p)
		}
	}
}

func TestInitialsFor(t *testing.T) {
	cases := map[string]string{
		"David Ochieng":        "DO",
		"a":                    "A",
		"":                     "U",
		"  ":                   "U",
		"Grace Wanjiru Mwangi": "GW",
	}
	for name, want := range cases {
		if got := initialsFor(name); got != want {
			t.Fatalf("initialsFor(%q) = %q, want %q", name, got, want)
		}
	}
}
