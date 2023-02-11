package apis

import (
	"github.com/gofiber/fiber/v2"
	. "treehole_backend/models"
)

// GetCurrentUser godoc
//
//	@Summary		get current user
//	@Tags			user
//	@Produce		json
//	@Router			/users/me [get]
//	@Success		200	{object}	User
//	@Failure		404	{object}	utils.MessageResponse	"User not found"
//	@Failure		500	{object}	utils.MessageResponse
func GetCurrentUser(c *fiber.Ctx) error {
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}
	var user User
	err = DB.Take(&user, userID).Error
	if err != nil {
		return err
	}
	return c.JSON(user)
}
