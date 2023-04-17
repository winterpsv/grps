package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"log"
	"task4_1/user-management/internal/controller/http/handler"
	"task4_1/user-management/internal/infrastructure/config"
	registry "task4_1/user-management/internal/infrastructure/registry/app"
	"task4_1/user-management/pkg/datastore"
)

func main() {
	cfg, err := config.ReadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	redis := datastore.NewRedisDB(cfg)
	db := datastore.NewMongoDB(cfg)

	r := registry.NewRegistry(db, redis, cfg)

	route := handler.NewRoute(echo.New(), r.NewAppControllers())
	route.InitRoutes()

	fmt.Println("Server listen at http://localhost" + ":" + cfg.ServerAddress)
	if err := route.Server.Start(":" + cfg.ServerAddress); err != nil {
		log.Fatalln(err)
	}
}
