package store

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestEncryptorRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	enc, err := newEncryptor(key)
	if err != nil {
		t.Fatal(err)
	}
	const plain = "mongodb://user:pass@host:27017/?replicaSet=rs0"

	ct, err := enc.seal(plain)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(ct, []byte("pass")) {
		t.Fatal("ciphertext leaks plaintext")
	}
	got, err := enc.open(ct)
	if err != nil {
		t.Fatal(err)
	}
	if got != plain {
		t.Fatalf("round trip mismatch: got %q want %q", got, plain)
	}
}

func TestEncryptorDisabledPassthrough(t *testing.T) {
	enc, err := newEncryptor(nil)
	if err != nil {
		t.Fatal(err)
	}
	if enc != nil {
		t.Fatal("expected nil encryptor when key is empty")
	}
	// nil encryptor must pass through plaintext.
	ct, err := enc.seal("plain")
	if err != nil {
		t.Fatal(err)
	}
	if string(ct) != "plain" {
		t.Fatalf("passthrough seal mismatch: %q", ct)
	}
	got, err := enc.open(ct)
	if err != nil {
		t.Fatal(err)
	}
	if got != "plain" {
		t.Fatalf("passthrough open mismatch: %q", got)
	}
}

func TestEncryptorOpenRejectsTampered(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)
	enc, _ := newEncryptor(key)
	ct, _ := enc.seal("secret")
	ct[len(ct)-1] ^= 0xFF // flip a bit in the ciphertext
	if _, err := enc.open(ct); err == nil {
		t.Fatal("expected auth failure on tampered ciphertext")
	}
}
