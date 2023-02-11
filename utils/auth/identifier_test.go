package auth

import (
	"testing"
)

func TestCheckPassword(t *testing.T) {
	rawPassword := "1234567890"
	encryptPassword := "pbkdf2_sha256$216000$dYxeEYraGSmj$QEOeBVq9oLuVS6T/vlpkzR7fMmAydKfP2SKo5XsiGOI="

	ok, err := CheckPassword(rawPassword, encryptPassword)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("error check password")
	}
}

func TestMakePassword(t *testing.T) {
	rawPassword := "1234567890"
	encryptPassword, err := MakePassword(rawPassword)
	if err != nil {
		t.Fatal(err)
	}
	println(encryptPassword)

	ok, err := CheckPassword(rawPassword, encryptPassword)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("error check password")
	}
}
