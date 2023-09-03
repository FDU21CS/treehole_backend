package models

import (
	"time"
	"treehole_backend/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/plugin/dbresolver"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Floor struct {
	/// saved fields
	ID        int       `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	/// base info

	// content of the floor, no more than 15000
	Content string `json:"content" gorm:"size:15000"`

	// a random username
	Anonyname string `json:"anonyname" gorm:"size:32"`

	// the ranking of this floor in the hole
	Ranking int `json:"ranking" gorm:"default:0;not null;uniqueIndex:idx_hole_ranking,priority:2"`

	// the modification times of floor.content
	Modified int `json:"modified" gorm:"not null;default:0"`

	// additional info, like "树洞管理团队"
	SpecialTag string `json:"special_tag"`

	Deleted bool `json:"deleted"`

	HoleID int `json:"hole_id" gorm:"uniqueIndex:idx_hole_ranking,priority:1"`

	// the user who wrote it
	UserID int `json:"-" gorm:"not null"`

	// a floor has many history
	History FloorHistorySlice `json:"history,omitempty"`

	/// dynamically generated fields

	// whether the user is the author of the floor
	IsMe bool `json:"is_me" gorm:"-:all"`
}

func (floor *Floor) GetID() int {
	return floor.ID
}

type Floors []*Floor

/******************************
Get and List
*******************************/

func (floor *Floor) Preprocess(c *fiber.Ctx) error {
	return Floors{floor}.Preprocess(c)
}

func (floors Floors) Preprocess(c *fiber.Ctx) error {
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	for i, floor := range floors {
		floors[i].IsMe = userID == floor.UserID
	}
	return nil
}

/******************************
Create
*******************************/

func (floor *Floor) Create(tx *gorm.DB) (err error) {
	// load floor mention, in another session
	var hole Hole

	err = tx.Clauses(dbresolver.Write).Transaction(func(tx *gorm.DB) error {
		// get anonymous name
		floor.Anonyname, err = FindOrGenerateAnonyname(tx, floor.HoleID, floor.UserID)
		if err != nil {
			return err
		}

		// get and lock hole for updating reply
		err = tx.Clauses(LockingClause).Take(&hole, floor.HoleID).Error
		if err != nil {
			return err
		}

		hole.Reply++
		floor.Ranking = hole.Reply

		// create floor, set floor_mention association in AfterCreate hook
		err = tx.Omit(clause.Associations).Create(&floor).Error
		if err != nil {
			return err
		}

		// update hole reply and update_at
		return tx.Model(&hole).
			Select("Reply").
			Updates(&hole).Error
	})
	if err != nil {
		return err
	}

	// delete cache
	return utils.DeleteCache(hole.CacheName())
}

// Backup Update and Modify
func (floor *Floor) Backup(tx *gorm.DB, userID int, reason string) error {
	history := FloorHistory{
		Content: floor.Content,
		FloorID: floor.ID,
		Reason:  reason,
		UserID:  userID,
	}
	return tx.Create(&history).Error
}
