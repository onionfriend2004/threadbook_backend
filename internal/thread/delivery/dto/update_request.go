package dto

type UpdateThreadRequest struct {
	ID       int     `json:"id"`
	EditorID int     `json:"editor_id"`
	Title    *string `json:"title"`
	Type     *string `json:"type"`
}
