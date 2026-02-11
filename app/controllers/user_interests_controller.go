package controller

import (
	"net/http"

	"group1-userservice/app/interfaces"
	"group1-userservice/app/middleware"

	"github.com/gin-gonic/gin"
)

type UserInterestsController struct {
	Service     interfaces.UserInterestsService
	UserService interfaces.UserService
}

func NewUserInterestsController(s interfaces.UserInterestsService, us interfaces.UserService) *UserInterestsController {
	return &UserInterestsController{Service: s, UserService: us}
}

// InterestsResponse is the API response format for interests
type InterestsResponse struct {
	Email     string                                `json:"email"`
	Interests []interfaces.UserInterestResponseItem `json:"interests"`
}

// UserInterestsUpdateRequest is the API request body for updating interests
type UserInterestsUpdateRequest struct {
	Interests []interfaces.UserInterestItemInput `json:"interests"`
}

// @Summary Get user interests
// @Description Get interests for the authenticated user
// @Tags Interests
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Success 200 {object} controller.InterestsResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/me/interests [get]
func (uc *UserInterestsController) GetForMe(c *gin.Context) {
	sub, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := uc.UserService.GetByKeycloakID(sub)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	email := user.Email

	items, err := uc.Service.GetForUser(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load interests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email":     email,
		"interests": items,
	})
}

// @Summary Update user interests
// @Description Update interests for the authenticated user (send array with {id,value})
// @Tags Interests
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param body body controller.UserInterestsUpdateRequest true "Updated interests"
// @Success 200 {object} controller.InterestsResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/me/interests [put]
func (uc *UserInterestsController) UpdateForMe(c *gin.Context) {
	sub, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := uc.UserService.GetByKeycloakID(sub)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input interfaces.UserInterestsUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	email := user.Email

	items, err := uc.Service.UpdateForUser(email, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email":     email,
		"interests": items,
	})
}

// @Summary Get user interests (internal)
// @Description Internal endpoint for other services (requires X-Service-Token)
// @Tags Interests
// @Produce json
// @Param X-Service-Token header string true "Service token"
// @Param email path string true "User Email"
// @Success 200 {object} controller.InterestsResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /internal/users/{email}/interests [get]
func (uc *UserInterestsController) GetForUserInternal(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing email"})
		return
	}

	items, err := uc.Service.GetForUser(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load interests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email":     email,
		"interests": items,
	})
}
