package request

import (
	"github.com/go-resty/resty/v2"
	"log"
)

var client *resty.Client

// Init takes a resty.Client
func Init(c *resty.Client) {
	client = c

	log.Println("Request package successfully initialized")
}
