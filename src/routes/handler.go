package routes

import (
	"bash06/goimg/src/router"
	"bash06/goimg/src/worker"

	"gorm.io/gorm"
)

type Handler struct {
	R  *router.Router
	Db *gorm.DB
	W  *worker.Worker
}

func InitHandler(r *router.Router) error {

}
