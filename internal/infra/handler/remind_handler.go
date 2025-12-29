package handler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/app"
	commonv1 "github.com/KasumiMercury/primind-remind-time-mgmt/internal/gen/common/v1"
	remindv1 "github.com/KasumiMercury/primind-remind-time-mgmt/internal/gen/remind/v1"
	pjson "github.com/KasumiMercury/primind-remind-time-mgmt/internal/proto"
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
	ctx := c.Request.Context()
	slog.InfoContext(ctx, "handling create remind request",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
	)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read request body", "error", err)
		respondProtoError(c, http.StatusBadRequest, "validation_error", "failed to read request body", "")

		return
	}

	var req remindv1.CreateRemindRequest
	if err := pjson.Unmarshal(body, &req); err != nil {
		slog.WarnContext(ctx, "request unmarshal failed",
			"error", err,
			"path", c.Request.URL.Path,
		)
		respondProtoError(c, http.StatusBadRequest, "validation_error", err.Error(), "")

		return
	}

	if err := pjson.Validate(&req); err != nil {
		slog.WarnContext(ctx, "request validation failed",
			"error", err,
			"path", c.Request.URL.Path,
		)
		respondProtoError(c, http.StatusBadRequest, "validation_error", err.Error(), "")

		return
	}

	devices := make([]app.DeviceInput, 0, len(req.Devices))
	for _, d := range req.Devices {
		devices = append(devices, app.DeviceInput{
			DeviceID: d.DeviceId,
			FCMToken: d.FcmToken,
		})
	}

	times := make([]time.Time, 0, len(req.Times))
	for _, t := range req.Times {
		times = append(times, t.AsTime())
	}

	input := app.CreateRemindInput{
		Times:    times,
		UserID:   req.UserId,
		Devices:  devices,
		TaskID:   req.TaskId,
		TaskType: taskTypeToString(req.TaskType),
	}

	output, err := h.useCase.CreateRemind(ctx, input)
	if err != nil {
		h.handleError(c, err)

		return
	}

	slog.InfoContext(ctx, "reminds created successfully",
		"task_id", req.TaskId,
		"count", output.Count,
	)
	respondProtoReminds(c, http.StatusCreated, output)
}

func (h *RemindHandler) GetRemindsByTimeRange(c *gin.Context) {
	ctx := c.Request.Context()
	slog.InfoContext(ctx, "handling get reminds by time range request",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
	)

	var req GetRemindsByTimeRangeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		slog.WarnContext(ctx, "request validation failed",
			"error", err,
			"path", c.Request.URL.Path,
		)
		respondProtoError(c, http.StatusBadRequest, "validation_error", err.Error(), "")

		return
	}

	input := app.GetRemindsByTimeRangeInput{
		Start: req.Start,
		End:   req.End,
	}

	output, err := h.useCase.GetRemindsByTimeRange(ctx, input)
	if err != nil {
		h.handleError(c, err)

		return
	}

	slog.InfoContext(ctx, "reminds retrieved successfully",
		"count", output.Count,
		"start", req.Start,
		"end", req.End,
	)
	respondProtoReminds(c, http.StatusOK, output)
}

func (h *RemindHandler) UpdateThrottled(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	slog.InfoContext(ctx, "handling update throttled request",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"remind_id", id,
	)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read request body", "error", err)
		respondProtoError(c, http.StatusBadRequest, "validation_error", "failed to read request body", "")

		return
	}

	var req remindv1.UpdateThrottledRequest
	if err := pjson.Unmarshal(body, &req); err != nil {
		slog.WarnContext(ctx, "request unmarshal failed",
			"error", err,
			"path", c.Request.URL.Path,
		)
		respondProtoError(c, http.StatusBadRequest, "validation_error", err.Error(), "")

		return
	}

	input := app.UpdateThrottledInput{
		ID:        id,
		Throttled: req.Throttled,
	}

	output, err := h.useCase.UpdateThrottled(ctx, input)
	if err != nil {
		h.handleError(c, err)

		return
	}

	slog.InfoContext(ctx, "throttled status updated successfully",
		"remind_id", output.ID,
		"throttled", output.Throttled,
	)
	respondProtoRemind(c, http.StatusOK, output)
}

func (h *RemindHandler) DeleteRemind(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	slog.InfoContext(ctx, "handling delete remind request",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"remind_id", id,
	)

	input := app.DeleteRemindInput{
		ID: id,
	}

	err := h.useCase.DeleteRemind(ctx, input)
	if err != nil {
		h.handleError(c, err)

		return
	}

	slog.InfoContext(ctx, "remind deleted successfully",
		"remind_id", id,
	)
	c.Status(http.StatusNoContent)
}

func (h *RemindHandler) CancelRemind(c *gin.Context) {
	ctx := c.Request.Context()
	slog.InfoContext(ctx, "handling cancel remind request",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
	)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read request body", "error", err)
		respondProtoError(c, http.StatusBadRequest, "validation_error", "failed to read request body", "")

		return
	}

	var req remindv1.CancelRemindRequest
	if err := pjson.Unmarshal(body, &req); err != nil {
		slog.WarnContext(ctx, "request unmarshal failed",
			"error", err,
			"path", c.Request.URL.Path,
		)
		respondProtoError(c, http.StatusBadRequest, "validation_error", err.Error(), "")

		return
	}

	if err := pjson.Validate(&req); err != nil {
		slog.WarnContext(ctx, "request validation failed",
			"error", err,
			"path", c.Request.URL.Path,
		)
		respondProtoError(c, http.StatusBadRequest, "validation_error", err.Error(), "")

		return
	}

	input := app.CancelRemindByTaskIDInput{
		TaskID: req.TaskId,
		UserID: req.UserId,
	}

	err = h.useCase.CancelRemindByTaskID(ctx, input)
	if err != nil {
		h.handleError(c, err)

		return
	}

	slog.InfoContext(ctx, "reminds canceled successfully",
		"task_id", req.TaskId,
		"user_id", req.UserId,
	)
	c.Status(http.StatusNoContent)
}

func (h *RemindHandler) handleError(c *gin.Context, err error) {
	var validationErr *app.ValidationError
	if errors.As(err, &validationErr) {
		respondProtoError(c, http.StatusBadRequest, "validation_error", validationErr.Message, validationErr.Field)

		return
	}

	if errors.Is(err, app.ErrNotFound) {
		respondProtoError(c, http.StatusNotFound, "not_found", "resource not found", "")

		return
	}

	respondProtoError(c, http.StatusInternalServerError, "internal_error", "an internal error occurred", "")
}

func (h *RemindHandler) RegisterRoutes(router *gin.RouterGroup) {
	reminds := router.Group("/reminds")
	{
		reminds.POST("", h.CreateRemind)
		reminds.GET("", h.GetRemindsByTimeRange)
		reminds.POST("/:id/throttled", h.UpdateThrottled)
		reminds.DELETE("/:id", h.DeleteRemind)
		reminds.POST("/cancel", h.CancelRemind)
	}
}

func respondProtoError(c *gin.Context, status int, errType, message, field string) {
	resp := &remindv1.ErrorResponse{
		Error:   errType,
		Message: message,
		Field:   field,
	}

	respBytes, err := pjson.Marshal(resp)
	if err != nil {
		c.Status(http.StatusInternalServerError)

		return
	}

	c.Data(status, "application/json", respBytes)
}

func respondProtoReminds(c *gin.Context, status int, output app.RemindsOutput) {
	reminds := make([]*remindv1.Remind, 0, len(output.Reminds))
	for _, r := range output.Reminds {
		reminds = append(reminds, toProtoRemind(r))
	}

	resp := &remindv1.RemindsResponse{
		Reminds: reminds,
		Count:   output.Count,
	}

	respBytes, err := pjson.Marshal(resp)
	if err != nil {
		c.Status(http.StatusInternalServerError)

		return
	}

	c.Data(status, "application/json", respBytes)
}

func respondProtoRemind(c *gin.Context, status int, output app.RemindOutput) {
	resp := &remindv1.RemindResponse{
		Remind: toProtoRemind(output),
	}

	respBytes, err := pjson.Marshal(resp)
	if err != nil {
		c.Status(http.StatusInternalServerError)

		return
	}

	c.Data(status, "application/json", respBytes)
}

func toProtoRemind(r app.RemindOutput) *remindv1.Remind {
	devices := make([]*remindv1.Device, 0, len(r.Devices))
	for _, d := range r.Devices {
		devices = append(devices, &remindv1.Device{
			DeviceId: d.DeviceID,
			FcmToken: d.FCMToken,
		})
	}

	return &remindv1.Remind{
		Id:               r.ID,
		Time:             timestamppb.New(r.Time),
		UserId:           r.UserID,
		Devices:          devices,
		TaskId:           r.TaskID,
		TaskType:         stringToTaskType(r.TaskType),
		Throttled:        r.Throttled,
		CreatedAt:        timestamppb.New(r.CreatedAt),
		UpdatedAt:        timestamppb.New(r.UpdatedAt),
		SlideWindowWidth: r.SlideWindowWidth,
	}
}

func taskTypeToString(t commonv1.TaskType) string {
	name := t.String()
	if strings.HasPrefix(name, "TASK_TYPE_") {
		return strings.ToLower(strings.TrimPrefix(name, "TASK_TYPE_"))
	}

	return strings.ToLower(name)
}

func stringToTaskType(s string) commonv1.TaskType {
	upper := "TASK_TYPE_" + strings.ToUpper(s)
	if v, ok := commonv1.TaskType_value[upper]; ok {
		return commonv1.TaskType(v)
	}

	return commonv1.TaskType_TASK_TYPE_UNSPECIFIED
}
