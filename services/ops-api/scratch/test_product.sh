#!/bin/bash
export EMAIL="stress_user_123@test.com"
export PASSWORD="password123"

curl -s -X POST http://localhost:8080/api/v1/onboard \
  -H "Content-Type: application/json" \
  -d '{"org_name": "Test Org", "tenant_name": "Test Tenant", "admin_email": "'$EMAIL'", "admin_password": "'$PASSWORD'"}' > /dev/null

ID_TOKEN=$(curl -s -X POST "http://shubhams-mac-mini.local:9099/identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=fake-api-key" \
  -H "Content-Type: application/json" \
  -d '{"email":"'$EMAIL'", "password":"'$PASSWORD'", "returnSecureToken":true}' | grep -o '"idToken": *"[^"]*"' | cut -d '"' -f 4)

curl -i -X POST http://localhost:8080/api/v1/pim/products \
  -H "Authorization: Bearer $ID_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "Test Product", "description": "Desc", "category": "Test"}'
