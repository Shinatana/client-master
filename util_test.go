package client

import (
	"encoding/base64"
	"fmt"
	"testing"
)

const (
	testUserName = "user"
	testUserPass = "pass"
)

func TestPrepareBasicAuth(t *testing.T) {
	hash := PrepareBasicAuth(testUserName, testUserPass)

	decodedBytes, _ := base64.StdEncoding.DecodeString(hash)
	decodedString := string(decodedBytes)

	if decodedString != fmt.Sprintf("%s:%s", testUserName, testUserPass) {
		t.Error("Decoded string is not equal source passed strings")
	}
}
