package controller

import (
	"fmt"
	"net/http"

	"group1-userservice/app/interfaces"
	"group1-userservice/app/models"
	"group1-userservice/app/service"
	"group1-userservice/app/storage"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserController struct {
	UserService  interfaces.UserService
	BadgeService interfaces.UserBadgeService
}

type UserPublicInfoResponse struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type UploadProfilePhotoResponse struct {
	Message   string `json:"message" example:"profile picture uploaded"`
	PublicURL string `json:"public_url" example:"http://localhost:9000/user-media/users/<sub>/profile.jpg"`
}

type PresignProfilePhotoGetResponse struct {
	URL string `json:"url"`
}

func NewUserController(us interfaces.UserService, bs interfaces.UserBadgeService) *UserController {
	return &UserController{
		UserService:  us,
		BadgeService: bs,
	}
}

// @Summary Get user email by firstname-lastname
// @Description Returns only first_name, last_name and email for the matched user. Requires Bearer access token.
// @Tags Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param firstname path string true "First name"
// @Param lastname path string true "Last name"
// @Success 200 {object} controller.UserPublicInfoResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/{firstname}/{lastname} [get]
func (uc *UserController) GetByFirstLast(c *gin.Context) {
	first := c.Param("firstname")
	last := c.Param("lastname")

	if first == "" || last == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing firstname or lastname"})
		return
	}

	info, err := uc.UserService.GetPublicInfoByFirstLast(first, last)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, info)

}

// @Summary Get user by email (internal)
// @Description Internal endpoint - requires X-Service-Token
// @Tags Users
// @Produce json
// @Param X-Service-Token header string true "Service token"
// @Param email path string true "User email"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/{email} [get]
func (uc *UserController) GetByEmail(c *gin.Context) {
	email := c.Param("email")

	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing email"})
		return
	}

	user, err := uc.UserService.GetByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, user)
}

// @Summary Get user by Keycloak subject
// @Description Get a single user by their Keycloak subject (sub) ID
// @Tags Users
// @Produce json
// @Param sub path string true "Keycloak subject ID"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string "Missing Keycloak sub"
// @Failure 404 {object} map[string]string "User not found"
// @Router /users/keycloak/{sub} [get]
func (uc *UserController) GetByKeycloakSub(c *gin.Context) {
	sub := c.Param("sub")

	if sub == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing Keycloak sub"})
		return
	}

	user, err := uc.UserService.GetByKeycloakID(sub)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, user)
}

// @Summary Update my user profile
// @Description Update logged-in user's profile data
// @Tags Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param request body models.User true "Updated user fields"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /users/me [put]
func (uc *UserController) UpdateMe(c *gin.Context) {
	userIDAny, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input models.UserUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	user, err := uc.UserService.GetByKeycloakID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	updated, err := uc.UserService.UpdateByEmail(user.Email, &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if service.IsProfileComplete(updated) {
		_, _ = uc.BadgeService.AwardBadge(updated.ID, service.BadgeKeyProfileComplete)
	}

	updated.Password = ""
	c.JSON(http.StatusOK, updated)
}

// @Summary Presign profile photo download
// @Description Returns a presigned GET URL to download the logged-in user's profile photo from MinIO/S3.
// @Tags Users
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Success 200 {object} controller.PresignProfilePhotoGetResponse
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User or profile photo not found"
// @Failure 500 {object} map[string]string "Server error"
// @Router /users/me/profile-photo/url [get]
func (uc *UserController) PresignProfilePhotoGet(s3 *storage.S3) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDAny, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		sub, ok := userIDAny.(string)
		if !ok || sub == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// get user to find profile photo URL
		user, err := uc.UserService.GetByKeycloakID(sub)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		if strings.TrimSpace(user.ProfilePhotoURL) == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "no profile photo"})
			return
		}

		// extract key from URL
		base := strings.TrimRight(s3.PublicBaseURL, "/") + "/"
		if !strings.HasPrefix(user.ProfilePhotoURL, base) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "profile_photo_url not compatible with S3_PUBLIC_BASE_URL"})
			return
		}

		key := strings.TrimPrefix(user.ProfilePhotoURL, base)

		signed, err := s3.PresignGet(key, 10*time.Minute)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, PresignProfilePhotoGetResponse{URL: signed})
	}
}

// @Summary Upload profile photo
// @Description Uploads a profile photo (multipart/form-data) to MinIO/S3, stores the public URL in the user profile, and returns the public URL.
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param file formData file true "Profile photo file (jpeg/png/webp). Max 5MB."
// @Success 200 {object} controller.UploadProfilePhotoResponse
// @Failure 400 {object} map[string]string "Missing file / unsupported type / too large"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /users/me/profile-photo [post]
func (uc *UserController) UploadProfilePhoto(s3 *storage.S3) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get user sub from context
		userIDAny, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		sub, ok := userIDAny.(string)
		if !ok || sub == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// get file from form
		fh, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing file field 'file'"})
			return
		}

		// basic size check
		const maxSize = 5 * 1024 * 1024
		if fh.Size > maxSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 5MB)"})
			return
		}

		// open file
		f, err := fh.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open upload"})
			return
		}
		defer f.Close()

		// content type check
		ct := fh.Header.Get("Content-Type")
		allowedCT := map[string]bool{
			"image/jpeg": true,
			"image/png":  true,
			"image/webp": true,
		}
		if !allowedCT[ct] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported content type"})
			return
		}

		// determine file extension
		ext := "jpg"
		if ct == "image/png" {
			ext = "png"
		} else if ct == "image/webp" {
			ext = "webp"
		} else if ct == "image/jpeg" {
			ext = "jpg"
		}

		key := fmt.Sprintf("users/%s/profile.%s", sub, ext)

		// upload to minio
		if err := s3.PutMultipart(key, f, fh.Size, ct); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// update user profile with public URL
		publicURL := strings.TrimRight(s3.PublicBaseURL, "/") + "/" + key
		if err := uc.UserService.UpdateProfilePhotoURLByKeycloakID(sub, publicURL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Load the full user so we have the User.ID (UUID)
		user, err := uc.UserService.GetByKeycloakID(sub)
		if err != nil {
			fmt.Println("WARN: failed to load user after profile photo upload:", err)
		} else {
			_, err := uc.BadgeService.AwardBadge(user.ID, service.BadgeKeyProfilePhotoUploaded)
			if err != nil {
				fmt.Println("WARN: failed to award profile_photo_uploaded badge:", err)
			}

			if service.IsProfileComplete(user) {
				awarded, err := uc.BadgeService.AwardBadge(user.ID, service.BadgeKeyProfileComplete)
				if err != nil {
					fmt.Println("WARN: failed to award complete_profile badge:", err)
				} else if awarded {
					fmt.Println("INFO: awarded badge 'profile_complete' to user", user.Email)
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "profile picture uploaded",
			"public_url": publicURL,
		})
	}
}

// @Summary Get my badges
// @Description Returns all badges earned by the authenticated user.
// @Description Each item now also includes `badge` metadata (key, name, description) in addition to the UserBadge fields.
// @Tags Badges
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Success 200 {array} models.UserBadge
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/me/badges [get]
func (uc *UserController) GetMyBadges(c *gin.Context) {
	subAny, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sub, ok := subAny.(string)
	if !ok || sub == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := uc.UserService.GetByKeycloakID(sub)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	badges, err := uc.BadgeService.GetBadgesForUser(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load badges"})
		return
	}

	c.JSON(http.StatusOK, badges)
}

// @Summary Get badges for a user by ID
// @Description Returns all badges earned by the given user. Requires Bearer access token.
// @Description Each item now also includes `badge` metadata (key, name, description) in addition to the UserBadge fields.
// @Tags Badges
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param id path string true "User ID (UUID)"
// @Success 200 {array} models.UserBadge
// @Failure 400 {object} map[string]string "Invalid user ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Could not load badges"
// @Router /users/id/{id}/badges [get]
func (uc *UserController) GetBadgesByUserID(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user id"})
		return
	}

	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if _, err := uc.UserService.GetByID(userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	badges, err := uc.BadgeService.GetBadgesForUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load badges"})
		return
	}

	c.JSON(http.StatusOK, badges)
}
