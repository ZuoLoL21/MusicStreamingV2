package consts

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
