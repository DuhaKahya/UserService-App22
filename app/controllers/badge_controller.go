package controller

import (
	"group1-userservice/app/interfaces"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BadgeController struct {
	BadgeService interfaces.UserBadgeService
	UserService  interfaces.UserService
}

type AwardBadgeRequest struct {
	UserID   string `json:"user_id"`
	BadgeKey string `json:"badge_key"`
}

func NewBadgeController(bs interfaces.UserBadgeService, us interfaces.UserService) *BadgeController {
	return &BadgeController{bs, us}
}

func (bc *BadgeController) Award(c *gin.Context) {
	var req AwardBadgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid user_id"})
		return
	}

	_, err = bc.BadgeService.AwardBadge(userID, req.BadgeKey)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "ok"})
}
