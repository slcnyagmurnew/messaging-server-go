package handler

import (
	"github.com/gin-gonic/gin"
	"messaging-server/internal/db"
	"messaging-server/internal/logger"
	"messaging-server/internal/scheduler"
	"net/http"
	"strings"
)

// ActionRequest represents the payload for start/stop actions.
type ActionRequest struct {
	Action string `json:"action"` // "start" or "stop"
}

// SchedulerControl godoc
// @Summary      Control scheduler
// @Description  Start or stop the message-sending scheduler based on the provided action
// @Tags         scheduler
// @Accept       json
// @Produce      json
// @Param        action  body  ActionRequest true  "Action to perform: 'start' or 'stop'"
// @Success      200     {object}  map[string]string
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /api/messaging [post]
func SchedulerControl(cron *scheduler.Scheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ActionRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		action := strings.ToLower(req.Action)
		switch action {
		case "start":
			status, code := cron.Start()
			c.JSON(code, gin.H{"status": status})
			return

		case "stop":
			// spawn the Stop() call in its own goroutine
			// to response http call immediately
			go func() {
				cron.Stop()
				logger.Sugar.Info("background scheduler.Stop() finished")
			}()
			c.JSON(http.StatusAccepted, gin.H{"status": "stop requested"})
			return

		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action, must be 'start' or 'stop'"})
			return
		}
	}
}

// ListMessages godoc
// @Summary      List messages sent
// @Description  Show the messages that sent
// @Tags         scheduler
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /list [get]
func ListMessages(c *gin.Context) {
	msgs, err := db.PostgresConnection.GetSentMessages() // get list of messages
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, msgs)
}
