package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CaptchaClaims struct {
	Answer string `json:"answer"`
	jwt.RegisteredClaims
}

type CaptchaChallenge struct {
	CaptchaID string `json:"captcha_id"` // This is the signed JWT token containing the answer
	Question  string `json:"question"`   // e.g. "5 + 9 = ?"
}

// GenerateCaptchaChallenge creates a random math problem and returns a CaptchaChallenge
func GenerateCaptchaChallenge() (*CaptchaChallenge, error) {
	// Generate two random numbers between 1 and 20
	n1, err := rand.Int(rand.Reader, big.NewInt(20))
	if err != nil {
		return nil, err
	}
	n2, err := rand.Int(rand.Reader, big.NewInt(20))
	if err != nil {
		return nil, err
	}

	num1 := n1.Int64() + 1
	num2 := n2.Int64() + 1
	answer := num1 + num2

	question := fmt.Sprintf("%d + %d = ?", num1, num2)
	answerStr := fmt.Sprintf("%d", answer)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default_secret"
	}

	// Create a token containing the answer, valid for 5 minutes
	claims := CaptchaClaims{
		Answer: answerStr,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return nil, err
	}

	return &CaptchaChallenge{
		CaptchaID: tokenStr,
		Question:  question,
	}, nil
}

// VerifyCaptcha checks if the provided answer is correct for the given captchaID
func VerifyCaptcha(captchaID, answer string) (bool, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default_secret"
	}

	token, err := jwt.ParseWithClaims(captchaID, &CaptchaClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return false, errors.New("captcha expired or invalid")
	}

	if claims, ok := token.Claims.(*CaptchaClaims); ok && token.Valid {
		if claims.Answer == answer {
			return true, nil
		}
		return false, nil
	}

	return false, errors.New("invalid captcha token")
}
