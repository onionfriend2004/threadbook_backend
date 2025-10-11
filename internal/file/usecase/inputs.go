package usecase

type GetFileInput struct {
	Filename string
}

type SaveFileInput struct {
	Filename    string
	Data        []byte
	ContentType string
}

type DeleteFileInput struct {
	Filename string
}
