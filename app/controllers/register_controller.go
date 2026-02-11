package controller

import (
	"group1-userservice/app/interfaces"
	"group1-userservice/app/models"

	"github.com/gin-gonic/gin"

	"group1-userservice/app/metrics"
	"time"
)

type RegisterController struct {
	UserService interfaces.UserService
}

func NewRegisterController(us interfaces.UserService) *RegisterController {
	return &RegisterController{
		UserService: us,
	}
}

// @Summary Register a new user
// @Description Create a new user in the database and Keycloak
// @Tags Users
// @Accept json
// @Produce json
// @Param request body models.User true "User info"
// @Success 201 {object} models.User
// @Router /users/register [post]
func (rc *RegisterController) Handle(c *gin.Context) {
	start := time.Now()
	defer func() {
		metrics.UserRequestDuration.Observe(time.Since(start).Seconds())
	}()

	metrics.UserRequestsTotal.Inc()

	var user models.User

	if err := c.BindJSON(&user); err != nil {
		metrics.UserRequestOutcomesTotal.WithLabelValues("bad_request").Inc()
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	if err := rc.UserService.Register(&user); err != nil {
		if err.Error() == "email already exists" {
			metrics.UserRequestOutcomesTotal.WithLabelValues("conflict").Inc()
			c.JSON(409, gin.H{"error": "Email already in use"})
			return
		}

		metrics.UserRequestOutcomesTotal.WithLabelValues("error").Inc()
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	metrics.UserRequestOutcomesTotal.WithLabelValues("success").Inc()

	user.Password = ""
	c.JSON(201, user)
}
