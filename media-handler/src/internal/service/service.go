package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"example.com/media-handler/src/config"
	"example.com/media-handler/src/internal/models"
	"example.com/media-handler/src/internal/repository"
	"github.com/google/uuid"
)

type MediaHandlerService struct {
	mediaHandlerRepository *repository.MediaHandlerRepository
	masterUrl              string
}

func New(mediaHandlerRepository *repository.MediaHandlerRepository, cfg *config.Config) *MediaHandlerService {
	return &MediaHandlerService{
		mediaHandlerRepository: mediaHandlerRepository,
		masterUrl:              fmt.Sprintf("%s:%d", cfg.SeaweedFS.MasterIp, cfg.SeaweedFS.MasterPort),
	}
}

func (m *MediaHandlerService) UpdateAvatar(file *os.File, fileName string) (uuid.UUID, error) {
	file.Seek(0, 0)
	media, err := m.assignFileToSeaweedFS(file, fileName)
	if err != nil {
		return uuid.Nil, err
	}
	slog.Info(media.FileId)
	return media.ID, nil
}

func (m *MediaHandlerService) UploadMedia(messageId uuid.UUID, file multipart.File, fileHeader *multipart.FileHeader) (err error) {
	media, err := m.assignFileToSeaweedFS(file, fileHeader.Filename)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	mf := models.MessageIdXFileId{
		MessageId: messageId,
		FileId:    media.ID,
	}
	bytes, err := json.Marshal(mf)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	err = m.mediaHandlerRepository.PublishInFileLoadedChannel(bytes)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	return nil
}

func (m *MediaHandlerService) GetMedia(id uuid.UUID) ([]byte, error) {
	fileId, volumeAddress, err := m.lookUpForFileIdAndVolumeAddress(id)
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
	fileId, volumeAddress, err := m.lookUpForFileIdAndVolumeAddress(id)
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

func (m *MediaHandlerService) lookUpForFileIdAndVolumeAddress(id uuid.UUID) (string, string, error) {
	media, err := m.mediaHandlerRepository.FindById(id)
	if err != nil {
		return "", "", err
	}
	volumeId := strings.Split(media.FileId, ",")[0]

	url, err := m.mediaHandlerRepository.GetVolumeIp(volumeId)
	if err == nil {
		return media.FileId, url, nil
	}

	lookupResponse := &models.SeaweedFSLookupResponse{}
	res, err := http.Get(fmt.Sprintf("http://%s/dir/lookup?volumeId=%s", m.masterUrl, volumeId))
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	json.NewDecoder(res.Body).Decode(lookupResponse)

	url = lookupResponse.Locations[0].Url
	err = m.mediaHandlerRepository.CacheVolumeIp(volumeId, url)
	if err != nil {
		return "", "", err
	}

	return media.FileId, url, nil
}

func (m *MediaHandlerService) assignFileToSeaweedFS(file io.Reader, fileName string) (*models.Media, error) {
	assignResponse := &models.SeaweedFSAssignResponse{}
	res, err := http.Get(fmt.Sprintf("http://%s/dir/assign", m.masterUrl))
	if err != nil {
		return nil, err
	}
	json.NewDecoder(res.Body).Decode(assignResponse)
	defer res.Body.Close()

	b := &bytes.Buffer{}
	w := multipart.NewWriter(b)
	form, err := w.CreateFormFile("file", fileName)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(form, file)
	if err != nil {
		return nil, err
	}
	w.Close()

	addr := fmt.Sprintf("http://%s/%s", assignResponse.Url, assignResponse.Fid)
	slog.Info(fmt.Sprintf("File URL: %v", addr))
	req, err := http.NewRequest("POST", addr, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	err = m.mediaHandlerRepository.CacheVolumeIp(strings.Split(assignResponse.Fid, ",")[0], assignResponse.Url)
	if err != nil {
		return nil, err
	}

	id := uuid.New()
	media := models.New(id, assignResponse.Fid)
	err = m.mediaHandlerRepository.Save(media)
	if err != nil {
		return nil, err
	}
	return media, nil
}
