package lib

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flappy-bird-server/model"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/golang-jwt/jwt"
)

type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func SetField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		return errors.New("Provided value type didn't match obj field type")
	}

	structFieldValue.Set(val)
	return nil
}

func ReadJsonFromBody(r *http.Request, w http.ResponseWriter, body any) error {
	if r.Method != http.MethodPost {
		return errors.New("method not allowed")
	}

	bodyByte, err := io.ReadAll(r.Body)
	if err != nil {
		return errors.New("failed to read request body")
	}
	defer r.Body.Close()

	if err := json.Unmarshal(bodyByte, &body); err != nil {
		return errors.New("invalid JSON format")
	}
	return nil
}

func WriteJson(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ErrorJson(w http.ResponseWriter, status int, message string, fileName string) error {
	if fileName != "" {
		ErrorLogger(message, fileName)
	}

	out, err := json.Marshal(map[string]interface{}{
		"message": message,
	})
	if err != nil {
		return err
	}

	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	return nil
}

func ErrorJsonWithCode(w http.ResponseWriter, err error, status ...int) {
	log.Println(err.Error())
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload jsonResponse
	payload.Error = true
	payload.Message = err.Error()

	WriteJson(w, statusCode, payload)
}

func ErrorLogger(newLine string, fileName string) {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(newLine)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func HashString(text string) string {
	hash := sha256.New()
	finalString := os.Getenv("SECRET") + text
	hash.Write([]byte(finalString))
	hashedBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashedBytes)
	return hashString
}

func GenerateToken(id string) (string, error) {
	var JWT_SECRET = []byte(os.Getenv("SECRET"))
	expirationTime := time.Now().Add(60 * 24 * 30 * time.Minute)
	claims := &model.TokenPayload{
		Id: id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JWT_SECRET)

	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func Stringify(data interface{}) string {
	jsonString, err := json.Marshal(data)
	if err != nil {
		log.Fatal("Something went wrong while stringify ", data)
	}
	return string(jsonString)
}
