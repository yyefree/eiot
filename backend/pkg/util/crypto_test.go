package util

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestHashPassword(t *testing.T) {
	hashed, err := HashPassword("test123")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hashed == "" {
		t.Fatal("HashPassword returned empty string")
	}
	if hashed == "test123" {
		t.Fatal("HashPassword did not hash")
	}
}

func TestCheckPassword(t *testing.T) {
	hashed, _ := HashPassword("correct")
	if !CheckPassword("correct", hashed) {
		t.Error("CheckPassword should return true for correct password")
	}
	if CheckPassword("wrong", hashed) {
		t.Error("CheckPassword should return false for wrong password")
	}
	if CheckPassword("", hashed) {
		t.Error("CheckPassword should return false for empty password")
	}
}

func TestGenerateJWTAndParse(t *testing.T) {
	secret := "test-secret"
	token, err := GenerateJWT(42, "admin", secret)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateJWT returned empty token")
	}

	claims, err := ParseJWT(token, secret)
	if err != nil {
		t.Fatalf("ParseJWT failed: %v", err)
	}
	if claims["uid"] != float64(42) {
		t.Errorf("uid = %v, want 42", claims["uid"])
	}
	if claims["role"] != "admin" {
		t.Errorf("role = %v, want admin", claims["role"])
	}
	if claims["iss"] != "eiot" {
		t.Errorf("iss = %v, want eiot", claims["iss"])
	}
}

func TestParseJWT_WrongSecret(t *testing.T) {
	token, _ := GenerateJWT(1, "user", "right-secret")
	_, err := ParseJWT(token, "wrong-secret")
	if err == nil {
		t.Error("ParseJWT should fail with wrong secret")
	}
}

func TestParseJWT_Expired(t *testing.T) {
	secret := "test-secret"
	// Create an expired token manually
	claims := map[string]interface{}{
		"uid":  1,
		"role": "user",
		"iat":  time.Now().Add(-2 * time.Hour).Unix(),
		"exp":  time.Now().Add(-1 * time.Hour).Unix(),
		"iss":  "eiot",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	tokenStr, _ := token.SignedString([]byte(secret))

	_, err := ParseJWT(tokenStr, secret)
	if err == nil {
		t.Error("ParseJWT should fail for expired token")
	}
}

func TestGenerateCode(t *testing.T) {
	for i := 0; i < 10; i++ {
		code := GenerateCode()
		if len(code) != 6 {
			t.Errorf("GenerateCode() length = %d, want 6", len(code))
		}
		for _, c := range code {
			if c < '0' || c > '9' {
				t.Errorf("GenerateCode() contains non-digit: %c", c)
			}
		}
	}
	// Two codes should not be the same (statistical)
	c1 := GenerateCode()
	c2 := GenerateCode()
	if c1 == c2 {
		// This is technically possible but extremely unlikely (1/1000000)
		t.Log("Warning: two consecutive codes matched (statistically improbable)")
	}
}
