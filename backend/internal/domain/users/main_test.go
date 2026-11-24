package users

import (
	"os"
	"testing"
)

// TestMain is the Go equivalent of suite-level setup and teardown. The current
// tests use in-memory stores, so no external fixture needs to be created.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
