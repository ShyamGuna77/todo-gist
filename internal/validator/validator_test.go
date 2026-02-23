package validator

import "testing"

func TestValidatorAddFieldError_DoesNotOverwrite(t *testing.T) {
	var v Validator
	v.AddFieldError("email", "first")
	v.AddFieldError("email", "second")
	if got := v.FieldErrors["email"]; got != "first" {
		t.Fatalf("expected first error to remain, got %q", got)
	}
}

func TestValidatorAddNonFieldError_Collected(t *testing.T) {
	var v Validator
	v.AddNonFieldError("one")
	v.AddNonFieldError("two")
	if len(v.NonFieldErrors) != 2 {
		t.Fatalf("expected 2 non-field errors, got %d", len(v.NonFieldErrors))
	}
}

func TestValidatorValid(t *testing.T) {
	tests := []struct {
		name string
		v    Validator
		want bool
	}{
		{"empty", Validator{}, true},
		{"field error", Validator{FieldErrors: map[string]string{"x": "y"}}, false},
		{"non-field error", Validator{NonFieldErrors: []string{"oops"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.Valid(); got != tt.want {
				t.Fatalf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotBlank(t *testing.T) {
	if NotBlank("   ") {
		t.Fatalf("expected NotBlank to be false for whitespace")
	}
	if !NotBlank("x") {
		t.Fatalf("expected NotBlank to be true for non-empty")
	}
}

func TestMaxChars(t *testing.T) {
	if !MaxChars("abc", 3) {
		t.Fatalf("expected MaxChars to be true")
	}
	if MaxChars("abcd", 3) {
		t.Fatalf("expected MaxChars to be false")
	}
}

func TestMinChars(t *testing.T) {
	if !MinChars("abcd", 4) {
		t.Fatalf("expected MinChars to be true")
	}
	if MinChars("abc", 4) {
		t.Fatalf("expected MinChars to be false")
	}
}

func TestMatchesAndEmailRX(t *testing.T) {
	if !Matches("a@b.com", EmailRX) {
		t.Fatalf("expected valid email to match")
	}
	if Matches("not-an-email", EmailRX) {
		t.Fatalf("expected invalid email not to match")
	}
}

func TestPermittedValue(t *testing.T) {
	if !PermittedValue(7, 1, 7, 365) {
		t.Fatalf("expected value to be permitted")
	}
	if PermittedValue(9, 1, 7, 365) {
		t.Fatalf("expected value to be not permitted")
	}
}
