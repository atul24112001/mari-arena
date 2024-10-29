package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
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

func WriteJson(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if status != http.StatusOK {
		w.WriteHeader(status)
	}
	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

func ErrorJson(w http.ResponseWriter, err error) error {
	var payload jsonResponse
	payload.Error = true
	payload.Message = err.Error()

	// statusCode, exists := CustomErrorType[err]
	// if exists {
	// return WriteJson(w, statusCode, payload)
	// }
	return WriteJson(w, http.StatusBadRequest, payload)
}

func ErrorJsonWithCode(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload jsonResponse
	payload.Error = true
	payload.Message = err.Error()

	return WriteJson(w, statusCode, payload)
}
