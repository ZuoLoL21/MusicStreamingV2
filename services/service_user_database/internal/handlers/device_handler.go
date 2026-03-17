package handlers

import (
	"backend/internal/consts"
	"backend/internal/di"
	sqlhandler "backend/sql/sqlc"
	"net/http"
	"time"

	libsdi "libs/di"
	libsmiddleware "libs/middleware"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type DeviceHandler struct {
	logger  *zap.Logger
	config  *di.Config
	returns *libsdi.ReturnManager
	db      consts.DB
}

func NewDeviceHandler(logger *zap.Logger, config *di.Config, returns *libsdi.ReturnManager, db consts.DB) *DeviceHandler {
	return &DeviceHandler{
		logger:  logger,
		config:  config,
		returns: returns,
		db:      db,
	}
}

type deviceResponse struct {
	UUID       string    `json:"uuid"`
	DeviceID   string    `json:"device_id"`
	DeviceName *string   `json:"device_name"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

func (h *DeviceHandler) GetDevices(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	devices, err := h.db.GetDevicesForUser(r.Context(), userUUID)
	if err != nil {
		logger.Error("failed to fetch devices", zap.Error(err))
		h.returns.ReturnError(w, "failed to fetch devices", http.StatusInternalServerError)
		return
	}

	response := make([]deviceResponse, len(devices))
	for i, d := range devices {
		var deviceName *string
		if d.DeviceName.Valid {
			deviceName = &d.DeviceName.String
		}
		response[i] = deviceResponse{
			UUID:       uuid.UUID(d.Uuid.Bytes).String(),
			DeviceID:   uuid.UUID(d.DeviceID.Bytes).String(),
			DeviceName: deviceName,
			CreatedAt:  d.CreatedAt.Time,
			LastUsedAt: d.LastUsedAt.Time,
			ExpiresAt:  d.ExpiresAt.Time,
		}
	}

	h.returns.ReturnJSON(w, response, http.StatusOK)
}

func (h *DeviceHandler) RevokeDevice(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	deviceIDStr := mux.Vars(r)["device_id"]
	if deviceIDStr == "" {
		h.returns.ReturnError(w, "device_id required", http.StatusBadRequest)
		return
	}

	deviceID, err := uuidToPgtype(deviceIDStr)
	if err != nil {
		h.returns.ReturnError(w, "invalid device_id format", http.StatusBadRequest)
		return
	}

	err = h.db.RevokeDevice(r.Context(), sqlhandler.RevokeDeviceParams{
		UserUuid: userUUID,
		DeviceID: deviceID,
	})
	if err != nil {
		logger.Error("failed to revoke device", zap.Error(err))
		h.returns.ReturnError(w, "failed to revoke device", http.StatusInternalServerError)
		return
	}

	logger.Info("device revoked",
		zap.String("user_uuid", uuid.UUID(userUUID.Bytes).String()),
		zap.String("device_id", deviceIDStr))

	w.WriteHeader(http.StatusNoContent)
}

func (h *DeviceHandler) RevokeAllDevices(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	err := h.db.RevokeAllDevicesForUser(r.Context(), userUUID)
	if err != nil {
		logger.Error("failed to revoke all devices", zap.Error(err))
		h.returns.ReturnError(w, "failed to revoke all devices", http.StatusInternalServerError)
		return
	}

	logger.Info("all devices revoked",
		zap.String("user_uuid", uuid.UUID(userUUID.Bytes).String()))

	w.WriteHeader(http.StatusNoContent)
}
