package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	FindByUsername(username string) (*User, error)
	FindByAccount(account string) (*User, error)
	Create(user *User) error
}

// 创建结构体 实现interface 使其继承UserRepository类型
type gormUserRepository struct {
	db *gorm.DB
}

// 创建 UserRepository 实例 后续任何实现该接口的结构体都可以替换掉gormUserRepository
// 因为结构体才是具体实现也就是插头 而接口是插座
func NewUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}
func (r *gormUserRepository) FindByUsername(username string) (*User, error) {
	var user User
	err := r.db.Where("username = ?", username).Take(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (r *gormUserRepository) Create(user *User) error {
	return r.db.Create(user).Error
}
func (r *gormUserRepository) FindByAccount(account string) (*User, error) {
	var user User
	err := r.db.Where("username = ? OR email = ? OR phone = ?", account, account, account).Take(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Token 缓存操作接口
type TokenRepository interface {
	// Redis SDK依赖context 所以必须传入
	SetUserToken(ctx context.Context, userID string, token string, duration time.Duration) error
	GetUserToken(ctx context.Context, userID string) (string, error)
	DeleteUserToken(ctx context.Context, userID string) error
}
type redisTokenRepository struct {
	rdb *redis.Client
}

func NewRedisTokenRepository(rdb *redis.Client) TokenRepository {
	return &redisTokenRepository{rdb: rdb}
}
func (r *redisTokenRepository) SetUserToken(ctx context.Context, userID string, token string, duration time.Duration) error {
	key := fmt.Sprintf("token:%s", userID)
	// 不论是set还是get返回的都是命令对象 封装好了符合Go设计风格的返回结果 直接调用其方法就行
	return r.rdb.Set(ctx, key, token, duration).Err()
}
func (r *redisTokenRepository) GetUserToken(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("token:%s", userID)
	return r.rdb.Get(ctx, key).Result()
}
func (r *redisTokenRepository) DeleteUserToken(ctx context.Context, userID string) error {
	key := fmt.Sprintf("token:%s", userID)
	return r.rdb.Del(ctx, key).Err()
}

// 证码缓存操作接口
type CodeRepository interface {
	SetCode(ctx context.Context, target string, code string, duration time.Duration) error
	GetCode(ctx context.Context, target string) (string, error)
	DeleteCode(ctx context.Context, target string) error
}
type redisCodeRepository struct {
	rdb *redis.Client
}

func NewRedisCodeRepository(rdb *redis.Client) CodeRepository {
	return &redisCodeRepository{rdb: rdb}
}
func (r *redisCodeRepository) SetCode(ctx context.Context, target string, code string, duration time.Duration) error {
	key := fmt.Sprintf("code:%s", target)
	return r.rdb.Set(ctx, key, code, duration).Err()
}
func (r *redisCodeRepository) GetCode(ctx context.Context, target string) (string, error) {
	key := fmt.Sprintf("code:%s", target)
	return r.rdb.Get(ctx, key).Result()
}
func (r *redisCodeRepository) DeleteCode(ctx context.Context, target string) error {
	key := fmt.Sprintf("code:%s", target)
	return r.rdb.Del(ctx, key).Err()
}
