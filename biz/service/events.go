package service

import (
	"context"
	"labs/local-transcoder/helper/factory"
	custdb "labs/local-transcoder/internal/db"
	"labs/local-transcoder/internal/hikvision"
	"labs/local-transcoder/internal/ome"
	"labs/local-transcoder/models/events"
)

type CommandServiceInterface interface {
	PtzCtrl(ctx context.Context, req *events.PtzCtrlRequest) error
	DeviceInfo(ctx context.Context, req *events.CommandRetrieveDeviceInfo) error
	AddCamera(ctx context.Context, req *events.CommandAddCameraInfo) error
	StartStream(ctx context.Context, req *events.CommandStartStreamInfo) error
	EndStream(ctx context.Context, req *events.CommandEndStreamInfo) error
}
type CommandService struct {
	db              *custdb.LayeredDb
	omeClient       ome.OmeClientInterface
	hikvisionClient hikvision.Client
}

func NewCommandService() CommandServiceInterface {
	return &CommandService{
		db:              custdb.Layered(),
		omeClient:       factory.Ome(),
		hikvisionClient: factory.Hikvision(),
	}
}

func (s *CommandService) PtzCtrl(ctx context.Context, req *events.PtzCtrlRequest) error {
	panic("unimplemented")
}

func (s *CommandService) AddCamera(ctx context.Context, req *events.CommandAddCameraInfo) error {
	panic("unimplemented")
}

func (s *CommandService) DeviceInfo(ctx context.Context, req *events.CommandRetrieveDeviceInfo) error {
	panic("unimplemented")
}

func (s *CommandService) StartStream(ctx context.Context, req *events.CommandStartStreamInfo) error {
	panic("unimplemented")
}

func (s *CommandService) EndStream(ctx context.Context, req *events.CommandEndStreamInfo) error {
	return nil
}
