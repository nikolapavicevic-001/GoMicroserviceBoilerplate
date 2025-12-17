package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/microserviceboilerplate/device/internal/domain"
	
	db "github.com/microserviceboilerplate/device/internal/adapter/repository/db"
)

// PostgresRepository implements DeviceRepository using PostgreSQL
type PostgresRepository struct {
	pool   *pgxpool.Pool
	queries *db.Queries
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(pool *pgxpool.Pool) domain.DeviceRepository {
	return &PostgresRepository{
		pool:    pool,
		queries: db.New(pool),
	}
}

// GetByID retrieves a device by ID
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.Device, error) {
	deviceUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid device ID: %w", err)
	}

	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(deviceUUID); err != nil {
		return nil, fmt.Errorf("failed to convert UUID: %w", err)
	}

	row, err := r.queries.GetDevice(ctx, pgUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	return toDomainDevice(&row), nil
}

// Create creates a new device
func (r *PostgresRepository) Create(ctx context.Context, d *domain.Device) (*domain.Device, error) {
	row, err := r.queries.CreateDevice(ctx, db.CreateDeviceParams{
		Name:   d.Name,
		Type:   d.Type,
		Status: d.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	return toDomainDevice(&row), nil
}

// Update updates an existing device
func (r *PostgresRepository) Update(ctx context.Context, d *domain.Device) (*domain.Device, error) {
	deviceUUID, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid device ID: %w", err)
	}

	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(deviceUUID); err != nil {
		return nil, fmt.Errorf("failed to convert UUID: %w", err)
	}

	row, err := r.queries.UpdateDevice(ctx, db.UpdateDeviceParams{
		ID:     pgUUID,
		Name:   d.Name,
		Type:   d.Type,
		Status: d.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	return toDomainDevice(&row), nil
}

// Delete removes a device by ID
func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	deviceUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid device ID: %w", err)
	}

	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(deviceUUID); err != nil {
		return fmt.Errorf("failed to convert UUID: %w", err)
	}

	err = r.queries.DeleteDevice(ctx, pgUUID)
	if err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	return nil
}

// List retrieves devices with pagination
func (r *PostgresRepository) List(ctx context.Context, limit, offset int32) ([]*domain.Device, error) {
	rows, err := r.queries.ListDevices(ctx, db.ListDevicesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	devices := make([]*domain.Device, len(rows))
	for i, row := range rows {
		devices[i] = toDomainDevice(&row)
	}

	return devices, nil
}

// Count returns the total number of devices
func (r *PostgresRepository) Count(ctx context.Context) (int32, error) {
	count, err := r.queries.CountDevices(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count devices: %w", err)
	}

	return int32(count), nil
}

// toDomainDevice converts sqlc model to domain entity
func toDomainDevice(row *db.Device) *domain.Device {
	var deviceID string
	if row.ID.Valid {
		deviceUUID, err := uuid.FromBytes(row.ID.Bytes[:])
		if err == nil {
			deviceID = deviceUUID.String()
		}
	}

	var createdAt, updatedAt time.Time
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		updatedAt = row.UpdatedAt.Time
	}

	return &domain.Device{
		ID:        deviceID,
		Name:      row.Name,
		Type:      row.Type,
		Status:    row.Status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

