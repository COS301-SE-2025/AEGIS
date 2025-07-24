package auditlog

import "github.com/gin-gonic/gin"

// MakeActor extracts auditlog.Actor from context
func MakeActor(c *gin.Context) Actor {
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	userEmail, _ := c.Get("email") // optional

	return Actor{
		ID:        toString(userID),
		Role:      toString(userRole),
		Email:     toString(userEmail),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
