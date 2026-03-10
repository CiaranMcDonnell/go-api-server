package utils

const (
	ContextKeyUserID = "user_id"
	CookieName       = "authToken"
	BearerPrefix     = "Bearer "

	DefaultPaginationLimit = 10
)

var AllowedPaginationLimits = map[int]bool{
	5:  true,
	10: true,
	20: true,
	50: true,
}
