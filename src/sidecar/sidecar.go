package sidecar

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/CE-Thesis-2023/ltd/src/internal/hikvision"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/reconciler"
	"github.com/CE-Thesis-2023/ltd/src/service"
	"go.uber.org/zap"
)

type HttpSidecar struct {
	server         *http.Server
	commandService *service.CommandService
	metadata       reconciler.Metadata
}

func NewHttpSidecar(commandService *service.CommandService, metadata reconciler.Metadata) *HttpSidecar {
	s := &HttpSidecar{
		commandService: commandService,
		metadata:       metadata,
	}
	s.init()
	return s
}

func (s *HttpSidecar) init() {
	s.server = &http.Server{
		Addr:        ":5600",
		ReadTimeout: 5 * time.Second,
		Handler:     s.newServeMux(),
	}
}

func (s *HttpSidecar) Start() error {
	logger.SInfo("Starting HTTP sidecar",
		zap.String("addr", s.server.Addr))
	if err := s.server.ListenAndServe(); err != nil {
		logger.SError("Failed to start HTTP sidecar",
			zap.Error(err))
	}
	return nil
}

func (s *HttpSidecar) Stop(ctx context.Context) error {
	logger.SInfo("Stopping HTTP sidecar")
	if err := s.server.Shutdown(ctx); err != nil {
		logger.SError("Failed to stop HTTP sidecar",
			zap.Error(err))
	}
	return nil
}

func (s *HttpSidecar) newServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/ptz/status", s.handlePtzStatus)
	mux.HandleFunc("/ptz/relative", s.handlePtzRelative)

	return mux
}

func (s *HttpSidecar) handlePtzStatus(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	cameraName := query.Get("name")
	if len(cameraName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	camera, err := s.metadata.GetCameraByName(cameraName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	info, err := s.commandService.DeviceInfo(r.Context(), camera)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(info)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(resp)
	w.Header().
		Add("Content-Type", "application/json")
}

func (s *HttpSidecar) handlePtzRelative(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	cameraName := query.Get("name")
	if len(cameraName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	camera, err := s.metadata.GetCameraByName(cameraName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var req hikvision.PTZCtrlRelativeRequest
	if err := json.
		NewDecoder(r.Body).
		Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.commandService.PTZRelative(r.Context(), camera, &req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
