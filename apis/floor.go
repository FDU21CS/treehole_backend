package apis

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	. "treehole_backend/models"
	. "treehole_backend/utils"
)

// ListFloorsInAHole
//
// @Summary List Floors In A Hole
// @Tags Floor
// @Produce	json
// @Router /holes/{hole_id}/floors [get]
// @Param hole_id path int true "hole id"
// @Param object query ListFloorRequest false "query"
// @Success	200 {array} Floor
func ListFloorsInAHole(c *fiber.Ctx) error {
	// validate
	holeID, err := c.ParamsInt("id")
	if err != nil {
		return err
	}

	var query ListFloorRequest
	err = ValidateQuery(c, &query)
	if err != nil {
		return err
	}

	// get floors
	var floors Floors
	result := DB.Unscoped().Limit(query.Size).
		Where("hole_id = ? and ranking >= ?", holeID, query.StartFloor).
		Find(&floors)
	if result.Error != nil {
		return result.Error
	}

	return Serialize(c, &floors)
}

// GetFloor
//
// @Summary Get A Floor
// @Tags Floor
// @Produce json
// @Router /floors/{id} [get]
// @Param id path int true "id"
// @Success 200 {object} Floor
// @Failure 404 {object} utils.MessageResponse
func GetFloor(c *fiber.Ctx) error {
	// validate floor id
	floorID, err := c.ParamsInt("id")
	if err != nil {
		return err
	}

	// get floor
	var floor Floor
	result := DB.First(&floor, floorID)
	if result.Error != nil {
		return result.Error
	}

	return Serialize(c, &floor)
}

// CreateFloor
//
// @Summary Create A Floor
// @Tags Floor
// @Produce json
// @Router /holes/{hole_id}/floors [post]
// @Param hole_id path int true "hole id"
// @Param json body CreateFloorRequest true "json"
// @Success 201 {object} Floor
func CreateFloor(c *fiber.Ctx) error {
	var body CreateFloorRequest
	err := ValidateBody(c, &body)
	if err != nil {
		return err
	}

	holeID, err := c.ParamsInt("id")
	if err != nil {
		return err
	}

	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	// create floor
	floor := Floor{
		HoleID:  holeID,
		UserID:  userID,
		Content: body.Content,
		IsMe:    true,
	}
	err = floor.Create(DB)
	if err != nil {
		return err
	}

	return c.Status(201).JSON(&floor)
}

// ModifyFloor
//
// @Summary Modify A Floor
// @Tags Floor
// @Produce application/json
// @Router /floors/{id} [put]
// @Param id path int true "id"
// @Param json body ModifyFloorRequest true "json"
// @Success 200 {object} Floor
// @Failure 404 {object} utils.MessageResponse
func ModifyFloor(c *fiber.Ctx) error {
	var body ModifyFloorRequest
	err := ValidateBody(c, &body)
	if err != nil {
		return err
	}

	user, err := GetUser(c)
	if err != nil {
		return err
	}

	// parse floor_id
	floorID, err := c.ParamsInt("id")
	if err != nil {
		return err
	}

	var floor Floor
	err = DB.Transaction(func(tx *gorm.DB) error {
		// load floor, lock for update
		err = tx.Clauses(LockingClause).Take(&floor, floorID).Error
		if err != nil {
			return err
		}

		if user.ID != floorID {
			return Forbidden()
		}

		if body.Content != nil {
			floor.Content = *body.Content
		}

		if body.SpecialTag != "" {
			if !user.IsAdmin {
				return Forbidden()
			}
			floor.SpecialTag = body.SpecialTag
		}

		return tx.Save(&floor).Error
	})
	if err != nil {
		return err
	}

	return Serialize(c, &floor)
}

// DeleteFloor
//
// @Summary Delete A Floor
// @Tags Floor
// @Produce application/json
// @Router /floors/{id} [delete]
// @Param id path int true "id"
// @Param json	body DeleteFloorRequest true "json"
// @Success	200 {object} Floor
// @Failure	404 {object} utils.MessageResponse
func DeleteFloor(c *fiber.Ctx) error {
	var body DeleteFloorRequest
	err := ValidateBody(c, &body)
	if err != nil {
		return err
	}

	floorID, err := c.ParamsInt("id")
	if err != nil {
		return err
	}

	// get user
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	var floor Floor
	err = DB.Transaction(func(tx *gorm.DB) error {

		result := tx.Clauses(LockingClause).Take(&floor, floorID)
		if result.Error != nil {
			return result.Error
		}

		// permission
		if !((user.ID == floor.UserID && !floor.Deleted) || user.IsAdmin) {
			return Forbidden()
		}

		err = floor.Backup(tx, user.ID, body.Reason)
		if err != nil {
			return err
		}

		floor.Deleted = true
		floor.Content = generateDeleteReason(body.Reason, user.ID == floor.UserID)
		return tx.Save(&floor).Error
	})
	if err != nil {
		return err
	}

	return Serialize(c, &floor)
}

func generateDeleteReason(reason string, isOwner bool) string {
	if isOwner {
		return "该内容被作者删除"
	}
	if reason == "" {
		reason = "违反社区规范"
	}
	return fmt.Sprintf("该内容因%s被删除", reason)
}
