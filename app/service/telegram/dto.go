package telegram

type GenericCallbackDTO struct {
	Type string `json:"type"`
}

type HistoryCallbackDTO struct {
	Type string `json:"type"`
	ID   int64  `json:"id"`
}
