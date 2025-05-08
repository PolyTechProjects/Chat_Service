package dto

type SeaweedFSAssignResponse struct {
	Count     int    `json:"count"`
	Fid       string `json:"fid"`
	Url       string `json:"url"`
	PublicUrl string `json:"publicUrl"`
}

type SeaweedFSLookupResponse struct {
	VolumeId  int             `json:"volumeId"`
	Locations []PublicUrlXUrl `json:"locations"`
}

type PublicUrlXUrl struct {
	PublicUrl string `json:"publicUrl"`
	Url       string `json:"url"`
}
