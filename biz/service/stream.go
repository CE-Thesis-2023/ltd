package service

type StreamManagementServiceInterface interface {
	MediaService() MediaServiceInterface
}

type StreamManagementService struct {
	mediaService MediaServiceInterface
}

func NewStreamManagementService() *StreamManagementService {
	return &StreamManagementService{
		mediaService: newMediaService(),
	}
}

func (s *StreamManagementService) MediaService() MediaServiceInterface {
	return s.mediaService
}
