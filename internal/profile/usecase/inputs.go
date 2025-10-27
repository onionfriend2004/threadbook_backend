package usecase

import (
	"mime/multipart"
)

type Avatar struct {
	File        multipart.File
	Size        int64
	Filename    string
	ContentType string
	Filetype    string
}

type UpdateProfileInput struct {
	UserID   int
	Nickname *string
	Avatar   *Avatar
}
