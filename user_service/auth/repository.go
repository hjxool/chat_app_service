package auth

import (
	"gorm.io/gorm"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	FindByUsername(username string) (*User, error)
	Create(user *User) error
}

// 创建结构体 实现interface 使其继承UserRepository类型
type gormUserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建 UserRepository 实例 后续任何实现该接口的结构体都可以替换掉gormUserRepository
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
