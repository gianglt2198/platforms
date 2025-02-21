package common

type KeyType string

const (
	KEY_AUTH_USER      KeyType = "x-authorization-key"
	KEY_CURRENT_TRAN   KeyType = "current-transaction-key"
	KEY_CORRELATION_ID KeyType = "correlation-id"
)
