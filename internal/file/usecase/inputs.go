package usecase

import "io"

type GetFileInput struct {
	Filename string
	Bucket   string
}

type SaveFile struct {
	File        io.Reader
	Size        int64
	Filename    string
	ContentType string
	UserID      string
	FileType    string
}

type DeleteFileInput struct {
	Filename string
}
