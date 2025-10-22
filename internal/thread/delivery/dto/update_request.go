package dto

type UpdateThreadRequest struct {
	ID       uint    `json:"id"`
	EditorID uint    `json:"editor_id"`
	Title    *string `json:"title"`
	Type     *string `json:"type"`
}
