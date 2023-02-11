//	@title			Moss Backend
//	@version		0.0.1
//	@description	Moss Backend

//	@contact.name	Maintainer Chen Ke
//	@contact.url	https://danxi.fduhole.com/about
//	@contact.email	dev@fduhole.com

//	@license.name	Apache 2.0
//	@license.url	https://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8000
//	@BasePath	/api

package main

import (
	"github.com/gofiber/fiber/v2"
	"log"
	"os"
	"os/signal"
	"syscall"
	"treehole_backend/apis"
	"treehole_backend/config"
	_ "treehole_backend/docs"
	"treehole_backend/middlewares"
	"treehole_backend/models"
	"treehole_backend/utils"
	"treehole_backend/utils/auth"
	"treehole_backend/utils/kong"
)

func main() {
	config.InitConfig()
	models.InitDB()
	utils.InitCache()
	auth.InitCache()

	// connect to kong
	err := kong.Ping()
	if err != nil {
		panic(err)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: utils.MyErrorHandler,
	})
	middlewares.RegisterMiddlewares(app)
	apis.RegisterRoutes(app)

	go func() {
		err = app.Listen("0.0.0.0:8000")
		if err != nil {
			log.Println(err)
		}
	}()

	interrupt := make(chan os.Signal, 1)

	// wait for CTRL-C interrupt
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interrupt

	// close app
	err = app.Shutdown()
	if err != nil {
		log.Println(err)
	}
}
