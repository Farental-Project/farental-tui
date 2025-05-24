package request

import (
	"farental/internal"
	"log"
)

var ctx *internal.AppCtx

func Init(appCtx *internal.AppCtx) {
	ctx = appCtx

	log.Println("Request package successfully initialized")
}
