package api

import (
	"RIP/internal/app/handler"
	"RIP/internal/app/repository"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Server start!")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()

	r.LoadHTMLGlob("../../templates/*")
	r.Static("/resources", "../../resources")

	r.GET("/TG", handler.GetChannels)
	r.GET("/channel/:id", handler.GetChannel)
	r.GET("/tg/:id", handler.GetTG)

	r.Run()

	log.Println("Server terminated!")
}
