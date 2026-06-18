package e2e

import (
	"commerce_modules/tests/e2e/harness"
	"fmt"
	"testing"
)

func TestDebug(t *testing.T) {
	h := harness.Setup(t)
	fmt.Println("BaseURL:", h.Config.BaseURL)
	fmt.Printf("Client: %#v\n", h.Client)
	fmt.Printf("Transport: %#v\n", h.Client.Transport)
}
