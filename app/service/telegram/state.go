package telegram

import "roflbeacon2/pkg/database"

type BotState struct {
	Stage string

	FenceParams database.CreateFenceParams
}
