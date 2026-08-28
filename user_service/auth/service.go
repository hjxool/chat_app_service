package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"
	"user_service/common"
	"user_service/logger"

	"github.com/redis/go-redis/v9"
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

// 生成6位数字验证码
func generateCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	// % 格式化开始标识 0 补位方式 6 最小字符数 原始数值转为字符串后的长度小于 6，就会触发左侧补 0 d 十进制整数格式
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// AuthService 认证业务逻辑接口
type AuthService interface {
	Login(ctx context.Context, req *LoginRequest) (string, error)
	Register(ctx context.Context, req *RegisterRequest) (*User, error)
	Logout(ctx context.Context, userID string) error
	SendVerifyCode(ctx context.Context, req *SendCodeRequest) error
	VerifyCode(ctx context.Context, req *VerifyCodeRequest) error
}
type authService struct {
	// ⚠️ 这里使用依赖倒置原则 不绑定固定的结构体 而是绑定抽象接口 以方便随时替换结构体
	repo      UserRepository
	tokenRepo TokenRepository
	codeRepo  CodeRepository
}

// NewAuthService 创建 AuthService 实例
func NewAuthService(repo UserRepository, tokenRepo TokenRepository) AuthService {
	return &authService{repo: repo, tokenRepo: tokenRepo}
}
func (s *authService) Login(ctx context.Context, req *LoginRequest) (string, error) {
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

	// 先查 Redis，有未过期 Token 直接返回，避免重复生成
	cachedToken, err := s.tokenRepo.GetUserToken(ctx, userIDStr)
	if err == nil && cachedToken != "" {
		return cachedToken, nil
	}
	// redis.Nil 表示 key 不存在（正常缓存未命中），其他错误降级处理
	if err != nil && !errors.Is(err, redis.Nil) {
		logger.Warn("读取缓存 Token 失败，降级处理", zap.Error(err))
	}

	// 生成 JWT Token
	token, err := GenerateToken(userIDStr, user.Username)
	if err != nil {
		return "", err
	}
	// redis 保存 Token，指定过期时间
	if err := s.tokenRepo.SetUserToken(ctx, userIDStr, token, 2*time.Hour); err != nil {
		logger.Warn("缓存 Token 失败", zap.Error(err))
	}
	return token, nil
}
func (s *authService) Register(ctx context.Context, req *RegisterRequest) (*User, error) {
	// 先验证码
	stored, err := s.codeRepo.GetCode(ctx, req.Target)
	if errors.Is(err, redis.Nil) {
		return nil, errors.New("验证码已过期")
	}
	if stored != req.Code {
		return nil, errors.New("验证码错误")
	}
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
	// 删除验证码
	s.codeRepo.DeleteCode(ctx, req.Target)
	return newUser, nil
}
func (s *authService) Logout(ctx context.Context, userID string) error {
	// 登出删除 Redis 中的 Token
	return s.tokenRepo.DeleteUserToken(ctx, userID)
}
func (s *authService) SendVerifyCode(ctx context.Context, req *SendCodeRequest) error {
	code, err := generateCode()
	if err != nil {
		return err
	}
	// 发送前先存 Redis，TTL = 5分钟
	if err = s.codeRepo.SetCode(ctx, req.Target, code, 5*time.Minute); err != nil {
		return err
	}
	// 根据类型路由到对应发送逻辑
	switch req.Type {
	case "sms":
		return common.SendSMS(req.Target, code)
	case "email":
		return common.SendEmail(req.Target, code)
	default:
		return errors.New("不支持的发送类型")
	}
}
func (s *authService) VerifyCode(ctx context.Context, req *VerifyCodeRequest) error {
	stored, err := s.codeRepo.GetCode(ctx, req.Target)
	if errors.Is(err, redis.Nil) {
		return errors.New("验证码已过期")
	}
	if err != nil {
		return err
	}
	if stored != req.Code {
		return errors.New("验证码错误")
	}
	// 验证成功立即删除，防止重放攻击
	s.codeRepo.DeleteCode(ctx, req.Target)
	return nil
}
