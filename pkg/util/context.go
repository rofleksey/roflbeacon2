package util

type ContextKey string

func (c ContextKey) String() string {
	return "roflbeacon2_" + string(c)
}

var UsernameContextKey ContextKey = "username"
var IpContextKey ContextKey = "ip"
