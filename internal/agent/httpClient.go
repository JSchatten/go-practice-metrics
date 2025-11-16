package agent

import (
	"time"

	"github.com/go-resty/resty/v2"
)

// Возможно отдельной переменной включать gZip?
// но это на будущее
func CreateHTTPClient() *resty.Client {
	client := resty.New()

	client.SetTimeout(10 * time.Second)

	// По заданию мы пишем свой ретрай, пока отключен
	// client.SetRetryCount(3)
	// client.SetRetryWaitTime(1 * time.Second)
	// client.SetRetryMaxWaitTime(3 * time.Second)

	return client
}
