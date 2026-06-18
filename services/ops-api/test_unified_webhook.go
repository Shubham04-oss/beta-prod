package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
)

func main() {
	secret := os.Getenv("UNIFIED_WEBHOOK_SECRET")
	if secret == "" {
		secret = "dummy_secret"
	}

	payload := []byte(`{"connection_id":"test_conn_123", "event":"item.created", "data": {"id":"abc"}}`)
	
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))

	req, _ := http.NewRequest("POST", "http://localhost:8080/unified/webhook", bytes.NewBuffer(payload))
	req.Header.Set("X-Unified-Signature", signature)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %v\n", resp.Status)
}
