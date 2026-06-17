package message

import "webscape/server/game/world"

const (
	QuestRewardDeliveryInventory = "inventory"
	QuestRewardDeliveryDropped   = "dropped"
)

type QuestRewardDelivery struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Count    int    `json:"count"`
	Delivery string `json:"delivery"`
}

type questCompletedData struct {
	QuestId              string                `json:"questId"`
	DisplayName          string                `json:"displayName,omitempty"`
	Description          string                `json:"description,omitempty"`
	CompletedStepId      string                `json:"completedStepId,omitempty"`
	CompletedStepSummary string                `json:"completedStepSummary,omitempty"`
	Rewards              []QuestRewardDelivery `json:"rewards"`
}

func NewQuestCompletedMessage(
	quest world.Quest,
	completedStep world.QuestStep,
	rewards []QuestRewardDelivery,
) Message {
	return newMessage(
		MessageTypeQuestCompleted,
		questCompletedData{
			QuestId:              quest.Id,
			DisplayName:          quest.DisplayName,
			Description:          quest.Description,
			CompletedStepId:      completedStep.Id,
			CompletedStepSummary: completedStep.Description,
			Rewards:              rewards,
		},
	)
}
