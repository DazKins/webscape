package message

import "webscape/server/game/world"

type questStartedData struct {
	QuestId            string `json:"questId"`
	DisplayName        string `json:"displayName,omitempty"`
	Description        string `json:"description,omitempty"`
	CurrentStepId      string `json:"currentStepId,omitempty"`
	CurrentStepSummary string `json:"currentStepSummary,omitempty"`
}

func NewQuestStartedMessage(quest world.Quest, currentStep world.QuestStep) Message {
	return newMessage(
		MessageTypeQuestStarted,
		questStartedData{
			QuestId:            quest.Id,
			DisplayName:        quest.DisplayName,
			Description:        quest.Description,
			CurrentStepId:      currentStep.Id,
			CurrentStepSummary: currentStep.Description,
		},
	)
}
