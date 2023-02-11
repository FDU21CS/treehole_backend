package apis

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	. "treehole_backend/models"
	. "treehole_backend/utils"
	"treehole_backend/utils/auth"
	"treehole_backend/utils/kong"
)

// Register godoc
//
//	@Summary		register
//	@Description	register with email or phone, password and verification code
//	@Tags			account
//	@Accept			json
//	@Produce		json
//	@Router			/register [post]
//	@Param			json	body		RegisterRequest	true	"json"
//	@Success		201		{object}	TokenResponse
//	@Failure		400		{object}	utils.MessageResponse	"验证码错误、用户已注册"
//	@Failure		500		{object}	utils.MessageResponse
func Register(c *fiber.Ctx) error {
	scope := "register"
	var body RegisterRequest
	err := ValidateBody(c, &body)
	if err != nil {
		return err
	}

	ok := auth.CheckVerificationCode(body.Email, scope, body.Verification)
	if !ok {
		return BadRequest("verification code error")
	}

	var user User

	err = DB.Take(&user, "email = ?", body.Email).Error
	// registered
	if err == nil {
		return BadRequest("您已注册，如果忘记密码，请使用忘记密码功能找回")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// registered
	user.Password, err = auth.MakePassword(body.Password)
	if err != nil {
		return err
	}

	err = DB.Create(&user).Error
	if err != nil {
		return err
	}

	err = kong.CreateUser(user.ID)
	if err != nil {
		return err
	}

	accessToken, refreshToken, err := kong.CreateToken(&user)
	if err != nil {
		return err
	}

	err = auth.DeleteVerificationCode(body.Email, scope)
	if err != nil {
		return err
	}

	return c.JSON(TokenResponse{
		Access:  accessToken,
		Refresh: refreshToken,
		Message: "register successful",
	})
}

// ChangePassword godoc
//
//	@Summary		reset password
//	@Description	reset password, reset jwt credential
//	@Tags			account
//	@Accept			json
//	@Produce		json
//	@Router			/register [put]
//	@Param			json	body		RegisterRequest	true	"json"
//	@Success		200		{object}	TokenResponse
//	@Failure		400		{object}	utils.MessageResponse	"验证码错误"
//	@Failure		500		{object}	utils.MessageResponse
func ChangePassword(c *fiber.Ctx) error {
	scope := "reset"
	var body RegisterRequest
	err := ValidateBody(c, &body)
	if err != nil {
		return err
	}

	ok := auth.CheckVerificationCode(body.Email, scope, body.Verification)
	if !ok {
		return BadRequest("验证码错误")
	}

	var user User
	err = DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("email = ?", body.Email).Take(&user).Error
		if err != nil {
			return err
		}

		user.Password, err = auth.MakePassword(body.Password)
		if err != nil {
			return err
		}
		return tx.Save(&user).Error
	})
	if err != nil {
		return err
	}

	err = kong.DeleteJwtCredential(user.ID)
	if err != nil {
		return err
	}

	accessToken, refreshToken, err := kong.CreateToken(&user)
	if err != nil {
		return err
	}

	err = auth.DeleteVerificationCode(body.Email, scope)
	if err != nil {
		return err
	}

	return c.JSON(TokenResponse{
		Access:  accessToken,
		Refresh: refreshToken,
		Message: "reset password successful",
	})
}

// VerifyWithEmail godoc
//
//	@Summary		verify with email in query
//	@Description	verify with email in query, Send verification email
//	@Tags			account
//	@Produce		json
//	@Router			/verify/email [get]
//	@Param			email	query		EmailModel	true	"email"
//	@Success		200		{object}	VerifyResponse
//	@Failure		400		{object}	utils.MessageResponse	"已注册“
func VerifyWithEmail(c *fiber.Ctx) error {
	var query EmailModel
	err := ValidateQuery(c, &query)
	if err != nil {
		return BadRequest("invalid email")
	}

	var (
		user  User
		scope string
	)

	err = DB.Take(&user, "email = ?", query.Email).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		scope = "register"
	} else {
		scope = "reset"
	}

	code, err := auth.SetVerificationCode(query.Email, scope)
	if err != nil {
		return err
	}

	err = SendCodeEmail(code, query.Email)
	if err != nil {
		return err
	}

	return c.JSON(VerifyResponse{
		Message: "验证邮件已发送，请查收\n如未收到，请检查邮件地址是否正确，检查垃圾箱，或重试",
		Scope:   scope,
	})
}
