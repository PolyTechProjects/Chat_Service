package controller

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"example.com/users/src/internal/client"
	"example.com/users/src/internal/dto"
	"example.com/users/src/internal/service"
	"github.com/google/uuid"
)

type UsersController struct {
	UsersService *service.UsersService
	authClient   *client.AuthGRPCClient
}

func New(UsersService *service.UsersService, authClient *client.AuthGRPCClient) *UsersController {
	return &UsersController{
		UsersService: UsersService,
		authClient:   authClient,
	}
}

func (c *UsersController) UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
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
	user, err := c.UsersService.UpdateUser(profileLink, dto)
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

func (c *UsersController) GetProfilesHandler(w http.ResponseWriter, r *http.Request) {
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
	if !params.Has("profileLink") && !params.Has("userId") {
		users, err := c.UsersService.GetUsers()
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
	} else if params.Has("profileLink") {
		profileLink := params.Get("profileLink")
		users, err := c.UsersService.GetUser(profileLink)
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
	} else if params.Has("userId") {
		userId, err := uuid.Parse(params.Get("userId"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		users, err := c.UsersService.GetUserById(userId)
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
