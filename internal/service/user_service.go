package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"tensechassignment/internal/cache"
	"tensechassignment/internal/model"
	"tensechassignment/internal/repository"
)

type UserService struct {
	repo     repository.UserRepository
	cache    cache.Cache
	cacheTTL time.Duration
}

func NewUserService(
	repo repository.UserRepository,
	cache cache.Cache,
	cacheTTL time.Duration,
) *UserService {
	return &UserService{
		repo:     repo,
		cache:    cache,
		cacheTTL: cacheTTL,
	}
}

var ErrInvalidStatus = errors.New("status must be 'active' or 'inactive'")

func (s *UserService) CreateUser(
	ctx context.Context,
	tenantID string,
	req model.CreateUserRequest,
) (*model.User, error) {
	status := strings.ToLower(strings.TrimSpace(req.Status))

	if status == "" {
		status = "active"
	}

	if status != "active" && status != "inactive" {
		return nil, ErrInvalidStatus
	}

	user := &model.User{
		ID:         uuid.NewString(),
		TenantID:   tenantID,
		Email:      strings.ToLower(strings.TrimSpace(req.Email)),
		FirstName:  strings.TrimSpace(req.FirstName),
		LastName:   strings.TrimSpace(req.LastName),
		Department: strings.TrimSpace(req.Department),
		Status:     status,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func userCacheKey(tenantID, userID string) string {
	return "user:" + tenantID + ":" + userID
}

func (s *UserService) GetUser(
	ctx context.Context,
	tenantID string,
	userID string,
) (*model.User, error) {
	key := userCacheKey(tenantID, userID)

	if s.cache != nil {
		value, err := s.cache.Get(ctx, key)

		if err == nil {
			var user model.User

			if json.Unmarshal([]byte(value), &user) == nil {
				return &user, nil
			}

			log.Println("invalid value found in redis cache")
		} else if !errors.Is(err, redis.Nil) {
			log.Println("redis get failed, using mongodb:", err)
		}
	}

	user, err := s.repo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		data, err := json.Marshal(user)
		if err == nil {
			if err := s.cache.Set(ctx, key, data, s.cacheTTL); err != nil {
				log.Println("redis set failed:", err)
			}
		}
	}

	return user, nil
}

func (s *UserService) ListUsers(
	ctx context.Context,
	tenantID string,
	search string,
	status string,
	page int64,
	limit int64,
) (*model.ListUsersResponse, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	search = strings.TrimSpace(search)
	status = strings.ToLower(strings.TrimSpace(status))

	users, total, err := s.repo.List(
		ctx,
		tenantID,
		search,
		status,
		page,
		limit,
	)
	if err != nil {
		return nil, err
	}

	if users == nil {
		users = []model.User{}
	}

	totalPages := (total + limit - 1) / limit

	if totalPages == 0 {
		totalPages = 1
	}

	return &model.ListUsersResponse{
		Users:      users,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}