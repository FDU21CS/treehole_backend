package models

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
	"treehole_backend/config"
	"treehole_backend/utils"
)

type Hole struct {
	/// saved fields
	ID        int       `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"index:,sort:desc"`
	UpdatedAt time.Time `json:"updated_at" gorm:"index:,sort:desc"`
	Deleted   time.Time `json:"-" gorm:"index"`

	/// base info

	// 回复量（即该洞下 floor 的数量 - 1）
	Reply int `json:"reply" gorm:"not null;default:0"`

	/// association info, should add foreign key

	// 洞主 id，管理员不可见
	UserID int `json:"-" gorm:"not null"`

	// 楼层列表
	Floors Floors `json:"-"`

	// 匿名映射表
	Mapping Users `json:"-" gorm:"many2many:anonyname_mapping"`

	/// generated field

	// 返回给前端的楼层列表，包括首楼、尾楼和预加载的前 n 个楼层
	HoleFloor struct {
		FirstFloor *Floor `json:"first_floor"` // 首楼
		LastFloor  *Floor `json:"last_floor"`  // 尾楼
		Floors     Floors `json:"prefetch"`    // 预加载的楼层
	} `json:"floors" gorm:"-:all"`
}

func (hole *Hole) GetID() int {
	return hole.ID
}

func (hole *Hole) CacheName() string {
	return fmt.Sprintf("hole_%d", hole.ID)
}

type Holes []*Hole

/**************
	get hole methods
 *******************/

const HoleCacheExpire = time.Minute * 10

func loadFloors(holes Holes) error {
	if len(holes) == 0 {
		return nil
	}

	var holeIDs []int
	for _, hole := range holes {
		holeIDs = append(holeIDs, hole.ID)
	}

	// load all floors with holeIDs and ranking < HoleFloorSize or the last floor
	// sorted by hole_id asc first and ranking asc second
	var floors Floors
	err := DB.
		Raw(
			// using file sort
			`SELECT * FROM (? UNION ?) f ORDER BY hole_id, ranking`,
			// use index(idx_hole_ranking), type range, use MRR
			DB.Model(&Floor{}).Where("hole_id in ? and ranking < ?", holeIDs, config.Config.HoleFloorSize),

			// UNION, remove duplications
			// use index(idx_hole_ranking), type eq_ref
			DB.Model(&Floor{}).Where(
				"(hole_id, ranking) in (?)",
				// use index(PRIMARY), type range
				DB.Model(&Hole{}).Select("id", "reply").Where("id in ?", holeIDs),
			),
		).Scan(&floors).Error
	if err != nil {
		return err
	}
	if len(floors) == 0 {
		return nil
	}

	/*
			Bind floors to hole.
			Note that floor is grouped by hole_id in hole_id asc order
		and hole is in random order, so we have to find hole_id those floors
		belong to both at the beginning and after floor group has changed.
			To bind, we use two pointers. Binding occurs when the floor's hole_id
		has changed, or when the floor is the last floor.
			The complexity is O(m*n), where m is the number of holes and
		n is the number of floors. Given that m is relatively small,
		the complexity is acceptable.
	*/
	var left, right int
	index := slices.IndexFunc(holes, func(hole *Hole) bool {
		return hole.ID == floors[0].HoleID
	})
	for _, floor := range floors {
		if floor.HoleID != holes[index].ID {
			holes[index].Floors = floors[left:right]
			left = right
			index = slices.IndexFunc(holes, func(hole *Hole) bool {
				return hole.ID == floor.HoleID
			})
		}
		right++
	}
	holes[index].Floors = floors[left:right]

	for _, hole := range holes {
		hole.SetHoleFloor()
	}

	return nil
}

func (hole *Hole) Preprocess(c *fiber.Ctx) error {
	return Holes{hole}.Preprocess(c)
}

func (holes Holes) Preprocess(_ *fiber.Ctx) error {
	notInCache := make(Holes, 0, len(holes))

	for _, hole := range holes {
		cachedHole := new(Hole)
		ok := utils.GetCache(hole.CacheName(), &cachedHole)
		if !ok {
			notInCache = append(notInCache, hole)
		} else {
			*hole = *cachedHole
		}
	}

	if len(notInCache) > 0 {
		err := UpdateHoleCache(notInCache)
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateHoleCache(holes Holes) error {
	err := loadFloors(holes)
	if err != nil {
		return err
	}

	for _, hole := range holes {
		err = utils.SetCache(hole.CacheName(), hole, HoleCacheExpire)
		if err != nil {
			return err
		}
	}
	return nil
}

/************************
	create and modify hole methods
 ************************/

func (hole *Hole) SetHoleFloor() {
	holeFloorSize := len(hole.Floors)
	if holeFloorSize == 0 {
		return
	}

	hole.HoleFloor.FirstFloor = hole.Floors[0]
	hole.HoleFloor.LastFloor = hole.Floors[holeFloorSize-1]
	if holeFloorSize <= config.Config.HoleFloorSize {
		hole.HoleFloor.Floors = hole.Floors
	} else {
		hole.HoleFloor.Floors = hole.Floors[0 : holeFloorSize-1]
	}
}

func (hole *Hole) Create(tx *gorm.DB) (err error) {
	err = tx.Transaction(func(tx *gorm.DB) error {
		// Create hole
		err = tx.Omit(clause.Associations).Create(&hole).Error
		if err != nil {
			return err
		}
		hole.Floors[0].HoleID = hole.ID

		// New anonyname
		hole.Floors[0].Anonyname, err = NewAnonyname(tx, hole.ID, hole.UserID)
		if err != nil {
			return err
		}

		// Create floor
		return tx.Create(&hole.Floors[0]).Error
	})
	// transaction commit here
	if err != nil {
		return err
	}

	// set hole.HoleFloor
	hole.SetHoleFloor()

	// store into cache
	return utils.SetCache(hole.CacheName(), hole, HoleCacheExpire)
}
