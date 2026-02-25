package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

var (
	Validator          = validator.New()
	CLOUDINARY_SECRET  = os.Getenv("CLOUDINARY_SECRET")
	CLOUDINARY_API_KEY = os.Getenv("CLOUDINARY_API_KEY ")
)

func ChangeFontToBase64(laptopPath string) (string, error) {
	path, err := filepath.Abs(laptopPath)
	if err != nil {
		return "", err
	}

	// Read the font file
	fontData, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// Encode the font file data to base64
	encodedFont := base64.StdEncoding.EncodeToString(fontData)

	// Print the base64 string
	return encodedFont, nil
}

func WriteJSON(w http.ResponseWriter, status int, payload any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, map[string]string{"error": err.Error()})
}

func ParseJSON(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("No body in this request")
	}

	return json.NewDecoder(r.Body).Decode(payload)
}

func ValidateJson(payload any) error {
	// Validate the payload
	if err := Validator.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		return errors
	}

	return nil
}

func VerifyPassword(encryptedPassword string, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(encryptedPassword), []byte(password)); err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}
