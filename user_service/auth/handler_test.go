package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type mockAuthService struct{}

func (m *mockAuthService) Login(req *LoginRequest) (string, error) {
	if req.Username == "admin" && req.Password == "123456" {
		return "mock-jwt-token", nil
	}
	return "", ErrUserNotFound
}

func (m *mockAuthService) Register(req *RegisterRequest) (*User, error) {
	if req.Username == "existing" {
		return nil, ErrUserAlreadyExists
	}
	return &User{
		ID:       100,
		Username: req.Username,
	}, nil
}

func TestRegisterHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAuthHandler(&mockAuthService{})

	r := gin.New()
	r.POST("/api/register", handler.RegisterHandler)

	reqBody := map[string]string{
		"username": "",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	head, ok := resp["head"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected head in response")
	}

	if int(head["code"].(float64)) != http.StatusBadRequest {
		t.Errorf("expected business code %d, got %v", http.StatusBadRequest, head["code"])
	}
}

func TestRegisterHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAuthHandler(&mockAuthService{})

	r := gin.New()
	r.POST("/api/register", handler.RegisterHandler)

	reqBody := map[string]string{
		"username": "newuser",
		"password": "password123",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	head := resp["head"].(map[string]interface{})
	if int(head["code"].(float64)) != http.StatusOK {
		t.Errorf("expected business code %d, got %v", http.StatusOK, head["code"])
	}
}
