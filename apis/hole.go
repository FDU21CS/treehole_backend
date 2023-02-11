package apis

import (
	"github.com/gofiber/fiber/v2"
	"time"
	. "treehole_backend/models"
	. "treehole_backend/utils"
)

// ListHoles
//
// @Summary Old API for Listing Holes
// @Tags Hole
// @Produce	json
// @Router /holes [get]
// @Param object query ListHoleRequest false "query"
// @Success 200 {array} Hole
func ListHoles(c *fiber.Ctx) error {
	var query ListHoleRequest
	err := ValidateQuery(c, &query)
	if err != nil {
		return err
	}

	if query.StartTime.IsZero() {
		query.StartTime = time.Now()
	}

	var holes Holes
	err = DB.Order(query.Order+" desc").Where("? < ?", query.Order, query.StartTime).Find(&holes).Error
	if err != nil {
		return err
	}

	return Serialize(c, &holes)
}

// CreateHole
//
// @Summary Create A Hole
// @Description Create a hole, create floor binding to it and set the name mapping
// @Tags Hole
// @Produce json
// @Router /holes [post]
// @Param division_id path int true "division id"
// @Param json body CreateHoleRequest true "json"
// @Success 201 {object} Hole
func CreateHole(c *fiber.Ctx) error {
	// validate body
	var body CreateHoleRequest
	err := ValidateBody(c, &body)
	if err != nil {
		return err
	}

	if len([]rune(body.Content)) > 15000 {
		return BadRequest("文本限制 15000 字")
	}

	// get user from auth
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	if !user.IsAdmin && body.SpecialTag != "" {
		return Forbidden()
	}

	hole := Hole{
		Floors: Floors{{UserID: user.ID, Content: body.Content, IsMe: true, SpecialTag: body.SpecialTag}},
		UserID: user.ID,
	}
	err = hole.Create(DB)
	if err != nil {
		return err
	}

	return c.Status(201).JSON(&hole)
}

// DeleteHole
//
// @Summary Delete A Hole
// @Description Hide a hole, but visible to admins. This may affect many floors, DO NOT ABUSE!!!
// @Tags Hole
// @Produce application/json
// @Router /holes/{id} [delete]
// @Param id path int true "id"
// @Success 204
// @Failure 404 {object} utils.MessageResponse
func DeleteHole(c *fiber.Ctx) error {
	// validate holeID
	holeID, err := c.ParamsInt("id")
	if err != nil {
		return err
	}

	// get user
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	// permission
	if !user.IsAdmin {
		return Forbidden()
	}

	hole := Hole{ID: holeID}
	err = DB.Delete(&hole).Error
	if err != nil {
		return err
	}

	err = DeleteCache(hole.CacheName())
	if err != nil {
		return err
	}

	return c.Status(204).JSON(nil)
}
