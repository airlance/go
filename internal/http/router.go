package http

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.Engine,
	groupHandler *GroupHandler,
	authMiddleware gin.HandlerFunc,
) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Groups routes
		groups := v1.Group("/groups")
		{
			// Public routes
			groups.GET("", groupHandler.ListGroups)
			groups.GET("/:id", groupHandler.GetGroup)
			groups.GET("/by-group-id/:group_id", groupHandler.GetGroupByGroupID)

			// Protected routes
			groups.POST("", authMiddleware, groupHandler.CreateGroup)
			groups.POST("/by-group-id/:group_id/update-counters", authMiddleware, groupHandler.UpdateGroupCounters)
			groups.DELETE("/by-group-id/:group_id", authMiddleware, groupHandler.DeleteGroup)
			groups.POST("/ips", authMiddleware, groupHandler.AddIP)
			groups.POST("/ips/batch", authMiddleware, groupHandler.AddIPs)
		}

		// IPs routes
		ips := v1.Group("/ips")
		{
			// Public routes
			ips.GET("/oldest", groupHandler.GetOldestIP)
		}

		// Scores routes
		scores := v1.Group("/scores")
		{
			scores.POST("/submit", authMiddleware, groupHandler.SubmitScore)
		}
	}
}
