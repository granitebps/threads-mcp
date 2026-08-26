package domain

import (
	"errors"
	"testing"
)

func TestProviderError(t *testing.T) {
	cause := errors.New("upstream detail")
	err := &ProviderError{Code: CodeNetworkFailure, Message: "public message", Cause: cause}
	if err.Error() != "public message" {
		t.Fatalf("Error() = %q", err.Error())
	}
	if !errors.Is(err, cause) {
		t.Fatal("ProviderError does not unwrap its cause")
	}
}
