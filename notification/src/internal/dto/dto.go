package dto

type BindDeviceToUserRequest struct {
	UserId      string
	DeviceToken string
}

type UnbindDeviceFromUserRequest struct {
	UserId      string
	DeviceToken string
}

type DeleteUserRequest struct {
	UserId string
}

type UpdateOldDeviceOnUserRequest struct {
	UserId         string
	OldDeviceToken string
	NewDeviceToken string
}
