package models

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"strings"
	"time"
	"treehole_backend/config"
	"treehole_backend/utils"
)

type User struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	JoinedTime time.Time `json:"joined_time" gorm:"autoCreateTime"`
	LastLogin  time.Time `json:"last_login" gorm:"autoUpdateTime"`
	IsAdmin    bool      `json:"is_admin"`
	Email      string    `json:"email" gorm:"size:128;index:,length:5"`
	Password   string    `json:"-" gorm:"size:128"`
}

type Users []*User

func GetUserID(c *fiber.Ctx) (int, error) {
	if config.Config.Mode == "dev" || config.Config.Mode == "test" {
		return 1, nil
	}

	id, err := strconv.Atoi(c.Get("X-Consumer-Username"))
	if err != nil {
		return 0, utils.Unauthorized("Unauthorized")
	}

	return id, nil
}

func GetUser(c *fiber.Ctx) (*User, error) {
	user := new(User)
	if config.Config.Mode == "dev" || config.Config.Mode == "test" {
		user.ID = 1
		user.IsAdmin = true
		return user, nil
	}

	// get id
	userID, err := GetUserID(c)
	if err != nil {
		return nil, err
	}

	err = DB.Take(&user, userID).Error
	return user, err
}

// parseJWT extracts and parse token
func parseJWT(token string) (Map, error) {
	if len(token) < 7 {
		return nil, errors.New("bearer token required")
	}

	payloads := strings.SplitN(token[7:], ".", 3) // extract "Bearer "
	if len(payloads) < 3 {
		return nil, errors.New("jwt token required")
	}

	// jwt encoding ignores padding, so RawStdEncoding should be used instead of StdEncoding
	payloadBytes, err := base64.RawStdEncoding.DecodeString(payloads[1]) // the middle one is payload
	if err != nil {
		return nil, err
	}

	var value Map
	err = json.Unmarshal(payloadBytes, &value)
	return value, err
}

func GetUserByRefreshToken(c *fiber.Ctx) (*User, error) {
	// get id
	userID, err := GetUserID(c)
	if err != nil {
		return nil, err
	}

	tokenString := c.Get("Authorization")
	if tokenString == "" { // token can be in either header or cookie
		tokenString = c.Cookies("refresh")
	}

	payload, err := parseJWT(tokenString)
	if err != nil {
		return nil, err
	}

	if tokenType, ok := payload["type"]; !ok || tokenType != "refresh" {
		return nil, utils.Unauthorized("refresh token invalid")
	}

	var user User
	err = DB.Take(&user, userID).Error
	return &user, err
}
