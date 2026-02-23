package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/form/v4"
)

func TestDecodeForm_DecodesFields(t *testing.T) {
	app := &Application{
		FormDecoder: form.NewDecoder(),
	}

	type dstStruct struct {
		Email   string `form:"email"`
		Expires int    `form:"expires"`
	}

	body := url.Values{}
	body.Set("email", "a@b.com")
	body.Set("expires", "7")

	r := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var dst dstStruct
	if err := app.DecodeForm(r, &dst); err != nil {
		t.Fatalf("DecodeForm() error = %v", err)
	}
	if dst.Email != "a@b.com" || dst.Expires != 7 {
		t.Fatalf("decoded dst = %+v, want email/expires set", dst)
	}
}

func TestDecodeForm_PanicsOnInvalidDecoderError(t *testing.T) {
	app := &Application{
		FormDecoder: form.NewDecoder(),
	}

	r := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader("a=b"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()

	// form decoder requires a pointer to a struct/map; passing a non-pointer should
	// trigger *form.InvalidDecoderError, which our helper treats as programmer error.
	var notPtr int
	_ = app.DecodeForm(r, notPtr)
}

func TestHumanDate(t *testing.T) {
	if got := humanDate(time.Time{}); got != "" {
		t.Fatalf("humanDate(zero) = %q, want empty string", got)
	}
	tt := time.Date(2026, time.February, 11, 10, 30, 0, 0, time.UTC)
	if got := humanDate(tt); got != "11 Feb 2026" {
		t.Fatalf("humanDate() = %q, want %q", got, "11 Feb 2026")
	}
}
