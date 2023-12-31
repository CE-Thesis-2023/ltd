package service

import custdb "labs/local-transcoder/internal/db"

type StreamManagementService struct {
	db *custdb.LayeredDb
}

func NewStreamManagementService() *StreamManagementService {
	return &StreamManagementService{
		db: custdb.Layered(),
	}
}
