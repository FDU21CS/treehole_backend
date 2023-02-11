package apis

import (
	"github.com/gofiber/fiber/v2"
	"treehole_backend/data"
)

// Index
//
//	@Produce	application/json
//	@Router		/ [get]
//	@Success	200	{object}	models.Map
func Index(c *fiber.Ctx) error {
	return c.Send(data.MetaFile)
}
