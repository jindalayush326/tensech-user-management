package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"tensechassignment/internal/middleware"
	"tensechassignment/internal/model"
	"tensechassignment/internal/repository"
	"tensechassignment/internal/service"
	"tensechassignment/internal/utils"
)

type UserController struct {
	service *service.UserService
}

func NewUserController(service *service.UserService) *UserController {
	return &UserController{
		service: service,
	}
}

func (uc *UserController) CreateUser(c *gin.Context) {
	var request model.CreateUserRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	tenantID := middleware.TenantID(c)

	user, err := uc.service.CreateUser(
		c.Request.Context(),
		tenantID,
		request,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidStatus):
			utils.Error(c, http.StatusBadRequest, err.Error())

		case errors.Is(err, repository.ErrDuplicateUser):
			utils.Error(c, http.StatusConflict, err.Error())

		default:
			utils.Error(c, http.StatusInternalServerError, "failed to create user")
		}

		return
	}

	utils.Success(c, http.StatusCreated, user)
}

func (uc *UserController) GetUser(c *gin.Context) {
	userID := c.Param("id")
	tenantID := middleware.TenantID(c)

	user, err := uc.service.GetUser(
		c.Request.Context(),
		tenantID,
		userID,
	)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			utils.Error(c, http.StatusNotFound, "user not found")
			return
		}

		utils.Error(c, http.StatusInternalServerError, "failed to fetch user")
		return
	}

	utils.Success(c, http.StatusOK, user)
}

func (uc *UserController) ListUsers(c *gin.Context) {
	tenantID := middleware.TenantID(c)

	search := strings.TrimSpace(c.Query("search"))
	status := c.Query("status")

	page, err := strconv.ParseInt(
		c.DefaultQuery("page", "1"),
		10,
		64,
	)
	if err != nil || page < 1 {
		utils.Error(c, http.StatusBadRequest, "page must be a positive integer")
		return
	}

	limit, err := strconv.ParseInt(
		c.DefaultQuery("limit", "10"),
		10,
		64,
	)
	if err != nil || limit < 1 || limit > 100 {
		utils.Error(c, http.StatusBadRequest, "limit must be between 1 and 100")
		return
	}

	if status != "" && status != "active" && status != "inactive" {
		utils.Error(c, http.StatusBadRequest, "status must be active or inactive")
		return
	}

	result, err := uc.service.ListUsers(
		c.Request.Context(),
		tenantID,
		search,
		status,
		page,
		limit,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to list users")
		return
	}

	utils.Success(c, http.StatusOK, result)
}