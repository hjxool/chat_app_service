package auth

import (
	"testing"

	"gorm.io/gorm"
)

type mockUserRepository struct {
	users map[string]*User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*User),
	}
}

func (m *mockUserRepository) FindByUsername(username string) (*User, error) {
	if u, ok := m.users[username]; ok {
		return u, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserRepository) Create(user *User) error {
	user.ID = uint64(len(m.users) + 1)
	m.users[user.Username] = user
	return nil
}

type mockTokenRepository struct {
}

func newMockTokenRepository() *mockTokenRepository {
	return &mockTokenRepository{}
}

func (m *mockTokenRepository) Save(token string, userID uint64, expiresAt int64) error {
	return nil
}

func (m *mockTokenRepository) FindByToken(token string) (uint64, error) {
	return 0, nil
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	repo := newMockUserRepository()
	service := NewAuthService(repo)

	// 1. 测试用户注册
	regReq := &RegisterRequest{
		Username: "testuser",
		Password: "password123",
	}

	user, err := service.Register(regReq)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", user.Username)
	}

	// 2. 测试重复注册
	_, err = service.Register(regReq)
	if err != ErrUserAlreadyExists {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}

	// 3. 测试正确登录
	loginReq := &LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	token, err := service.Login(loginReq)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	if token == "" {
		t.Errorf("expected non-empty token")
	}

	// 4. 测试错误密码登录
	wrongLoginReq := &LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	_, err = service.Login(wrongLoginReq)
	if err != ErrPasswordIncorrect {
		t.Errorf("expected ErrPasswordIncorrect, got %v", err)
	}
}
