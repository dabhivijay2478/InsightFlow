package crypto

import "testing"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	master := "01234567890123456789012345678901234567890123456789012345678901"
	in := "super-secret-password-!@#"
	enc, err := Encrypt(master, in)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Decrypt(master, enc)
	if err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Fatalf("got %q want %q", out, in)
	}
}
