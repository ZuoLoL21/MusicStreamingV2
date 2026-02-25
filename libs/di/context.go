package di

type ContextKey string

const (
	RequestIDKey  ContextKey = "requestId"
	UserUUIDKey   ContextKey = "userUuid"
	ServiceJWTKey ContextKey = "serviceJwt"
)
