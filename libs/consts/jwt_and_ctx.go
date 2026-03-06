package consts

import "time"

type ContextKey string

const (
	RequestIDKey  ContextKey = "requestId"
	UserUUIDKey   ContextKey = "userUuid"
	ServiceJWTKey ContextKey = "serviceJwt"
)

const (
	JWTSubjectNormal  = "normal"
	JWTSubjectRefresh = "refresh"
	JWTSubjectService = "service"
)

const (
	JWTExpirationNormal  = 10 * time.Minute
	JWTExpirationRefresh = 10 * 24 * time.Hour // 10 days
	JWTExpirationService = 2 * time.Minute
)
