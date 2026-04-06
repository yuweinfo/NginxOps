package service

import (
	"errors"
	"log"
	"nginxops/internal/model"
	"nginxops/internal/repository"
	"nginxops/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService() *AuthService {
	return &AuthService{
		userRepo: repository.NewUserRepository(),
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrDatabaseNotConnected) {
			return nil, errors.New("系统初始化中，请稍后重试")
		}
		return nil, errors.New("用户名或密码错误")
	}
	if user == nil {
		return nil, errors.New("用户名或密码错误")
	}

	if !user.Enabled {
		return nil, errors.New("用户名或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	token, err := jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, errors.New("生成令牌失败")
	}

	return &LoginResponse{
		Token:    token,
		Username: user.Username,
		Role:     user.Role,
	}, nil
}

func (s *AuthService) InitDefaultUser() error {
	count, err := s.userRepo.Count()
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	// 创建默认管理员账号
	// 注意：通过 Welcome 引导页面配置的管理员账号会在此之前创建
	// 这里只是为了兼容性，确保在没有用户时有一个默认账号
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.User{
		Username: "admin",
		Password: string(hashedPassword),
		Email:    "admin@nginxops.local",
		Role:     "admin",
		Enabled:  true,
	}

	log.Printf("Creating default admin user: admin")
	return s.userRepo.Create(user)
}
