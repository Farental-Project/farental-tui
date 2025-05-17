package request

import (
	"farentalapp/data"
	"log"
)

var ctx *data.AppCtx

func Init(appCtx *data.AppCtx) {
	ctx = appCtx

	log.Println("Request package successfully initialized")
}
