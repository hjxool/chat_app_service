package auth

import (
	"context"
	"errors"
	"strconv"
	"time"
	"user_service/logger"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// go中枚举基本就是这么写的
var (
	ErrUserNotFound      = errors.New("用户名或密码错误")
	ErrPasswordIncorrect = errors.New("用户名或密码错误")
	ErrUserAlreadyExists = errors.New("用户名已存在")
)

// AuthService 认证业务逻辑接口
type AuthService interface {
	Login(req *LoginRequest) (string, error)
	Register(req *RegisterRequest) (*User, error)
}

type authService struct {
	// ⚠️ 这里使用依赖倒置原则 不绑定固定的结构体 而是绑定抽象接口 以方便随时替换结构体
	repo      UserRepository
	tokenRepo TokenRepository
}

// NewAuthService 创建 AuthService 实例
func NewAuthService(repo UserRepository, tokenRepo TokenRepository) AuthService {
	return &authService{repo: repo, tokenRepo: tokenRepo}
}

func (s *authService) Login(req *LoginRequest) (string, error) {
	user, err := s.repo.FindByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", ErrPasswordIncorrect
	}

	// strconv 用于将基本数据类型与string相互转换 FormatUint 是将无符号整数转换成10进制字符串
	userIDStr := strconv.FormatUint(user.ID, 10)
	// 生成 JWT Token
	token, err := GenerateToken(userIDStr, user.Username)
	if err != nil {
		return "", err
	}
	// redis 保存 Token，指定过期时间
	if err := s.tokenRepo.SetUserToken(context.Background(), userIDStr, token, 2*time.Hour); err != nil {
		logger.Warn("缓存 Token 失败", zap.Error(err))
	}
	return token, nil
}

func (s *authService) Register(req *RegisterRequest) (*User, error) {
	existingUser, err := s.repo.FindByUsername(req.Username)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 对密码进行 bcrypt 哈希加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := &User{
		Username: req.Username,
		Password: string(hashedPassword),
	}

	if err := s.repo.Create(newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}
