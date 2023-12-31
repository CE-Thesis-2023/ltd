package service

import custdb "labs/local-transcoder/internal/db"

type CommandService struct {
	db *custdb.LayeredDb
}

func NewCommandService() *CommandService {
	return &CommandService{
		db:          custdb.Layered(),
	}
}
