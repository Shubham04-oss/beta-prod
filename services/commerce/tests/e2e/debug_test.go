package e2e

import (
	"commerce_modules/tests/e2e/harness"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestDebug2(t *testing.T) {
	h := harness.Setup(t)
	req, _ := http.NewRequest("POST", h.Config.BaseURL+"/inventory/adjust", nil)
	resp, err := h.Client.Do(req)
	fmt.Println("Error:", err)
	if resp != nil {
		fmt.Println("StatusCode:", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Body:", string(body))
	}
}
