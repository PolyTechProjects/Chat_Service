package service

import (
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strings"

	"example.com/media/src/internal/client"
	"example.com/media/src/internal/models"
	"example.com/media/src/internal/repository"
	"github.com/google/uuid"
)

type MediaHandlerService struct {
	mediaHandlerRepository *repository.MediaHandlerRepository
	redisClient            *client.RedisClient
	seaweedFSClient        *client.SeaweedFSClient
}

func New(mediaHandlerRepository *repository.MediaHandlerRepository, redisClient *client.RedisClient, seaweedFSCLient *client.SeaweedFSClient) *MediaHandlerService {
	return &MediaHandlerService{
		mediaHandlerRepository: mediaHandlerRepository,
		redisClient:            redisClient,
		seaweedFSClient:        seaweedFSCLient,
	}
}

func (m *MediaHandlerService) UploadMedia(file multipart.File, fileHeader *multipart.FileHeader) error {
	assignResponse, err := m.seaweedFSClient.AssignAndUpload(file, fileHeader.Filename)
	if err != nil {
		return err
	}
	err = m.redisClient.CacheVolumeIp(strings.Split(assignResponse.Fid, ",")[0], assignResponse.Url)
	if err != nil {
		return err
	}

	id := uuid.New()
	media := models.NewMedia(id, assignResponse.Fid)
	err = m.mediaHandlerRepository.Save(media)
	if err != nil {
		return err
	}
	return nil
}

func (m *MediaHandlerService) GetMedia(id uuid.UUID) ([]byte, error) {
	fileId, volumeAddress, err := m.findVolumeAddress(id)
	if err != nil {
		return nil, err
	}
	slog.Info(fileId)
	slog.Info(volumeAddress)
	res, err := http.Get(fmt.Sprintf("http://%s/%s", volumeAddress, fileId))
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (m *MediaHandlerService) DeleteMedia(id uuid.UUID) error {
	fileId, volumeAddress, err := m.findVolumeAddress(id)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://%s/%s", volumeAddress, fileId), nil)
	if err != nil {
		return err
	}
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	err = m.mediaHandlerRepository.DeleteById(id)
	if err != nil {
		return err
	}
	return nil
}

func (m *MediaHandlerService) findVolumeAddress(id uuid.UUID) (string, string, error) {
	media, err := m.mediaHandlerRepository.FindById(id)
	if err != nil {
		return "", "", err
	}
	volumeId := strings.Split(media.FileId, ",")[0]

	url, err := m.redisClient.GetVolumeIp(volumeId)
	if err == nil {
		return media.FileId, url, nil
	}

	lookupResponse, err := m.seaweedFSClient.Lookup(volumeId)
	if err != nil {
		return "", "", err
	}

	url = lookupResponse.Locations[0].Url
	err = m.redisClient.CacheVolumeIp(volumeId, url)
	if err != nil {
		return "", "", err
	}

	return media.FileId, url, nil
}
