package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/microserviceboilerplate/device/internal/service"
	pb "github.com/microserviceboilerplate/device/proto/device"
)

// Handler handles NATS request-reply messages
type Handler struct {
	deviceService *service.DeviceService
	nc            *nats.Conn
}

// NewHandler creates a new NATS handler
func NewHandler(nc *nats.Conn, deviceService *service.DeviceService) *Handler {
	return &Handler{
		deviceService: deviceService,
		nc:            nc,
	}
}

// RegisterHandlers registers all NATS request handlers
func (h *Handler) RegisterHandlers() error {
	// List devices handler
	_, err := h.nc.Subscribe("request.device.list", h.handleListDevices)
	if err != nil {
		return fmt.Errorf("failed to subscribe to request.device.list: %w", err)
	}

	// Get device handler
	_, err = h.nc.Subscribe("request.device.get", h.handleGetDevice)
	if err != nil {
		return fmt.Errorf("failed to subscribe to request.device.get: %w", err)
	}

	return nil
}

// handleListDevices handles device list requests
func (h *Handler) handleListDevices(msg *nats.Msg) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req struct {
		Page     int32 `json:"page"`
		PageSize int32 `json:"page_size"`
	}

	if len(msg.Data) > 0 {
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			h.sendError(msg.Reply, "invalid request", err)
			return
		}
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	devices, total, err := h.deviceService.ListDevices(ctx, req.Page, req.PageSize)
	if err != nil {
		h.sendError(msg.Reply, "failed to list devices", err)
		return
	}

	// Convert to proto format for consistency
	protoDevices := make([]*pb.Device, len(devices))
	for i, d := range devices {
		protoDevices[i] = &pb.Device{
			Id:        d.ID,
			Name:      d.Name,
			Type:      d.Type,
			Status:    d.Status,
			CreatedAt: d.CreatedAt.Unix(),
			UpdatedAt: d.UpdatedAt.Unix(),
		}
	}

	response := &pb.ListDevicesResponse{
		Devices:  protoDevices,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	data, err := json.Marshal(response)
	if err != nil {
		h.sendError(msg.Reply, "failed to marshal response", err)
		return
	}

	if msg.Reply != "" {
		msg.Respond(data)
	}
}

// handleGetDevice handles single device requests
func (h *Handler) handleGetDevice(msg *nats.Msg) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.sendError(msg.Reply, "invalid request", err)
		return
	}

	if req.ID == "" {
		h.sendError(msg.Reply, "device ID is required", fmt.Errorf("empty ID"))
		return
	}

	device, err := h.deviceService.GetDevice(ctx, req.ID)
	if err != nil {
		h.sendError(msg.Reply, "failed to get device", err)
		return
	}

	protoDevice := &pb.Device{
		Id:        device.ID,
		Name:      device.Name,
		Type:      device.Type,
		Status:    device.Status,
		CreatedAt: device.CreatedAt.Unix(),
		UpdatedAt: device.UpdatedAt.Unix(),
	}

	data, err := json.Marshal(protoDevice)
	if err != nil {
		h.sendError(msg.Reply, "failed to marshal response", err)
		return
	}

	if msg.Reply != "" {
		msg.Respond(data)
	}
}

// sendError sends an error response
func (h *Handler) sendError(reply string, message string, err error) {
	if reply == "" {
		return
	}

	errorResp := map[string]interface{}{
		"error":   message,
		"details": err.Error(),
	}

	data, _ := json.Marshal(errorResp)
	h.nc.Publish(reply, data)
}

