package controller

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"example.com/user_mgmt/src/internal/client"
	"example.com/user_mgmt/src/internal/dto"
	"example.com/user_mgmt/src/internal/service"
)

type UserMgmtController struct {
	userMgmtService *service.UserMgmtService
	authClient      *client.AuthGRPCClient
}

func New(userMgmtService *service.UserMgmtService, authClient *client.AuthGRPCClient) *UserMgmtController {
	return &UserMgmtController{
		userMgmtService: userMgmtService,
		authClient:      authClient,
	}
}

func (c *UserMgmtController) UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	_, err := c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	dto := &dto.UpdateUserRequest{}
	err = json.NewDecoder(r.Body).Decode(dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	path := r.URL.Path
	segments := strings.Split(path, "/")
	profileLink := segments[len(segments)-1]
	user, err := c.userMgmtService.UpdateUser(profileLink, dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(resp)
}

func (c *UserMgmtController) GetProfilesHandler(w http.ResponseWriter, r *http.Request) {
	_, err := c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !params.Has("profileLink") {
		users, err := c.userMgmtService.GetUsers()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp, err := json.Marshal(users)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(resp)
	} else {
		profileLink := params.Get("profileLink")
		users, err := c.userMgmtService.GetUser(profileLink)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp, err := json.Marshal(users)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(resp)
	}
}
