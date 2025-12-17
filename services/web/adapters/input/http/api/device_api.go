package api

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	natsclient "github.com/microserviceboilerplate/web/adapters/output/nats"
	"github.com/microserviceboilerplate/web/domain/port/input"
	"github.com/microserviceboilerplate/web/domain/port/output"
	deviceservice "github.com/microserviceboilerplate/web/domain/service"
)

// DeviceAPIHandler handles device API requests
type DeviceAPIHandler struct {
	authUseCase   input.AuthUseCase
	deviceService *deviceservice.DeviceService
	templateRepo  output.TemplateRepository
}

// NewDeviceAPIHandler creates a new device API handler
func NewDeviceAPIHandler(authUseCase input.AuthUseCase, deviceService *deviceservice.DeviceService, templateRepo output.TemplateRepository) *DeviceAPIHandler {
	return &DeviceAPIHandler{
		authUseCase:   authUseCase,
		deviceService: deviceService,
		templateRepo:  templateRepo,
	}
}

// RegisterRoutes registers device API routes
func (h *DeviceAPIHandler) RegisterRoutes(r chi.Router) {
	r.Get("/devices", h.ListDevices)
	r.Get("/devices/create-form", h.CreateForm)
	r.Post("/devices", h.CreateDevice)
	r.Put("/devices/{id}", h.UpdateDevice)
	r.Delete("/devices/{id}", h.DeleteDevice)
}

// ListDevices returns devices table HTML
func (h *DeviceAPIHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), "request", r)
	ctx = context.WithValue(ctx, "response", w)

	// Check authentication
	exists, _ := h.authUseCase.ValidateSession(ctx, "")
	if !exists {
		errorHTML := h.renderErrorHTML("Unauthorized. Please log in.")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, errorHTML)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 10
	}

	// Call service - returns *natsclient.ListDevicesResponse
	resp, err := h.deviceService.ListDevices(ctx, int32(page), int32(pageSize))
	if err != nil {
		// Return error as HTML fragment for HTMX
		errorMsg := "Failed to load devices"
		if err != nil {
			errorMsg += ": " + err.Error()
		}
		errorHTML := h.renderErrorHTML(errorMsg)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, errorHTML)
		return
	}

	// Convert to HTML table
	tableHTML := h.renderDevicesTableHTML(resp)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, tableHTML)
}

func (h *DeviceAPIHandler) renderDevicesTableHTML(resp *natsclient.ListDevicesResponse) string {
	if resp == nil || len(resp.Devices) == 0 {
		return `<table><thead><tr><th>ID</th><th>Name</th><th>Type</th><th>Status</th><th>Created At</th><th>Actions</th></tr></thead><tbody><tr><td colspan="6" class="empty-state">No devices found</td></tr></tbody></table>`
	}

	htmlStr := `<table><thead><tr><th>ID</th><th>Name</th><th>Type</th><th>Status</th><th>Created At</th><th>Actions</th></tr></thead><tbody>`
	for _, device := range resp.Devices {
		createdAt := time.Unix(device.CreatedAt, 0).Format("2006-01-02 15:04:05")
		// Escape all user-generated content to prevent XSS
		deviceID := html.EscapeString(device.ID)
		deviceName := html.EscapeString(device.Name)
		deviceType := html.EscapeString(device.Type)
		deviceStatus := html.EscapeString(device.Status)
		
		// Status is safe to use in class name as it comes from our system
		htmlStr += fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td>%s</td><td><span class="status-badge status-%s">%s</span></td><td>%s</td><td><button class="action-btn btn-edit" hx-get="/devices/%s/edit" hx-target="#device-modal-content">Edit</button><button class="action-btn btn-delete" hx-delete="/devices/%s" hx-target="#devices-table" hx-swap="innerHTML" hx-confirm="Are you sure you want to delete this device?">Delete</button></td></tr>`, 
			deviceID, deviceName, deviceType, deviceStatus, deviceStatus, html.EscapeString(createdAt), deviceID, deviceID)
	}
	htmlStr += `</tbody></table>`
	return htmlStr
}

func (h *DeviceAPIHandler) renderErrorHTML(message string) string {
	return fmt.Sprintf(`<div class="empty-state" style="padding: 40px; text-align: center; color: #ef4444;">%s</div>`, html.EscapeString(message))
}

// CreateForm returns device creation form HTML
func (h *DeviceAPIHandler) CreateForm(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), "request", r)
	ctx = context.WithValue(ctx, "response", w)

	// Check authentication
	exists, _ := h.authUseCase.ValidateSession(ctx, "")
	if !exists {
		errorHTML := h.renderErrorHTML("Unauthorized. Please log in.")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, errorHTML)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<form hx-post="/devices" hx-target="#devices-table" hx-swap="innerHTML" hx-trigger="submit">
		<div class="form-group">
			<label>Name</label>
			<input type="text" name="name" required style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;">
		</div>
		<div class="form-group">
			<label>Type</label>
			<input type="text" name="type" required style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;">
		</div>
		<div class="form-group">
			<label>Status</label>
			<select name="status" style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;">
				<option value="active">Active</option>
				<option value="inactive">Inactive</option>
			</select>
		</div>
		<button type="submit" style="background: #3b82f6; color: white; padding: 10px 20px; border: none; border-radius: 6px; cursor: pointer;">Create</button>
	</form>`)
}

// CreateDevice creates a new device
func (h *DeviceAPIHandler) CreateDevice(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), "request", r)
	ctx = context.WithValue(ctx, "response", w)

	// Check authentication
	exists, _ := h.authUseCase.ValidateSession(ctx, "")
	if !exists {
		errorHTML := h.renderErrorHTML("Unauthorized. Please log in.")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, errorHTML)
		return
	}

	w.Header().Set("HX-Trigger", "devicesUpdated")
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// UpdateDevice updates a device
func (h *DeviceAPIHandler) UpdateDevice(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), "request", r)
	ctx = context.WithValue(ctx, "response", w)

	// Check authentication
	exists, _ := h.authUseCase.ValidateSession(ctx, "")
	if !exists {
		errorHTML := h.renderErrorHTML("Unauthorized. Please log in.")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, errorHTML)
		return
	}

	w.Header().Set("HX-Trigger", "devicesUpdated")
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// DeleteDevice deletes a device
func (h *DeviceAPIHandler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), "request", r)
	ctx = context.WithValue(ctx, "response", w)

	// Check authentication
	exists, _ := h.authUseCase.ValidateSession(ctx, "")
	if !exists {
		errorHTML := h.renderErrorHTML("Unauthorized. Please log in.")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, errorHTML)
		return
	}

	w.Header().Set("HX-Trigger", "devicesUpdated")
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
