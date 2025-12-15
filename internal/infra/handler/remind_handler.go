package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/app"
)

type RemindHandler struct {
	useCase app.RemindUseCase
}

func NewRemindHandler(useCase app.RemindUseCase) *RemindHandler {
	return &RemindHandler{
		useCase: useCase,
	}
}

func (h *RemindHandler) CreateRemind(c *gin.Context) {
	slog.Info("handling create remind request",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
	)

	var req CreateRemindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("request validation failed",
			"error", err,
			"path", c.Request.URL.Path,
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
			Field:   "",
		})

		return
	}

	devices := make([]app.DeviceInput, 0, len(req.Devices))
	for _, d := range req.Devices {
		devices = append(devices, app.DeviceInput{
			DeviceID: d.DeviceID,
			FCMToken: d.FCMToken,
		})
	}

	input := app.CreateRemindInput{
		Times:    req.Times,
		UserID:   req.UserID,
		Devices:  devices,
		TaskID:   req.TaskID,
		TaskType: req.TaskType,
	}

	output, err := h.useCase.CreateRemind(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)

		return
	}

	slog.Info("reminds created successfully",
		"task_id", req.TaskID,
		"count", output.Count,
	)
	c.JSON(http.StatusCreated, FromDTOs(output))
}

func (h *RemindHandler) GetRemindsByTimeRange(c *gin.Context) {
	slog.Info("handling get reminds by time range request",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
	)

	var req GetRemindsByTimeRangeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		slog.Warn("request validation failed",
			"error", err,
			"path", c.Request.URL.Path,
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
			Field:   "",
		})

		return
	}

	input := app.GetRemindsByTimeRangeInput{
		Start: req.Start,
		End:   req.End,
	}

	output, err := h.useCase.GetRemindsByTimeRange(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)

		return
	}

	slog.Info("reminds retrieved successfully",
		"count", output.Count,
		"start", req.Start,
		"end", req.End,
	)
	c.JSON(http.StatusOK, FromDTOs(output))
}

func (h *RemindHandler) UpdateThrottled(c *gin.Context) {
	id := c.Param("id")

	slog.Info("handling update throttled request",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"remind_id", id,
	)

	var req UpdateThrottledRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("request validation failed",
			"error", err,
			"path", c.Request.URL.Path,
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
			Field:   "",
		})

		return
	}

	input := app.UpdateThrottledInput{
		ID:        id,
		Throttled: req.Throttled,
	}

	output, err := h.useCase.UpdateThrottled(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)

		return
	}

	slog.Info("throttled status updated successfully",
		"remind_id", output.ID,
		"throttled", output.Throttled,
	)
	c.JSON(http.StatusOK, FromDTO(output))
}

func (h *RemindHandler) DeleteRemind(c *gin.Context) {
	id := c.Param("id")

	slog.Info("handling delete remind request",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"remind_id", id,
	)

	input := app.DeleteRemindInput{
		ID: id,
	}

	err := h.useCase.DeleteRemind(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)

		return
	}

	slog.Info("remind deleted successfully",
		"remind_id", id,
	)
	c.Status(http.StatusNoContent)
}

func (h *RemindHandler) handleError(c *gin.Context, err error) {
	var validationErr *app.ValidationError
	if errors.As(err, &validationErr) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: validationErr.Message,
			Field:   validationErr.Field,
		})

		return
	}

	if errors.Is(err, app.ErrNotFound) {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "resource not found",
			Field:   "",
		})

		return
	}

	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   "internal_error",
		Message: "an internal error occurred",
		Field:   "",
	})
}

func (h *RemindHandler) RegisterRoutes(router *gin.RouterGroup) {
	reminds := router.Group("/reminds")
	{
		reminds.POST("", h.CreateRemind)
		reminds.GET("", h.GetRemindsByTimeRange)
		reminds.POST("/:id/throttled", h.UpdateThrottled)
		reminds.DELETE("/:id", h.DeleteRemind)
	}
}
