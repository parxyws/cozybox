package contextkey

type ContextKey string

const (
	RequestID ContextKey = "request_id"
	UserID    ContextKey = "user_id"
)
