package paymentsvc

import "time"

const (
	linkRetryLimit           = 2
	linkRetrySleep           = 12 * time.Second
	statusPollLimit          = 12
	qrisStatusPollLimit      = 300
	midtransChargeRetryLimit = 3
)
