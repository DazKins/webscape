package message

import "webscape/server/game/world"

type worldData struct {
	SizeX    int               `json:"sizeX"`
	SizeY    int               `json:"sizeY"`
	Terrain  []string          `json:"terrain"`
	Blockers [][]bool          `json:"blockers"`
	Walls    []world.WorldWall `json:"walls"`
	Quests   []questData       `json:"quests"`
}

type questData struct {
	Id           string           `json:"id"`
	DisplayName  string           `json:"displayName,omitempty"`
	Description  string           `json:"description,omitempty"`
	StartEventId string           `json:"startEventId,omitempty"`
	Steps        []questStepData  `json:"steps"`
	Rewards      questRewardsData `json:"rewards"`
}

type questStepData struct {
	Id          string               `json:"id"`
	Description string               `json:"description"`
	Requirement questRequirementData `json:"requirement"`
}

type questRequirementData struct {
	EventId string `json:"eventId"`
	Count   int    `json:"count"`
}

type questRewardsData struct {
	Items []questRewardItemData `json:"items"`
}

type questRewardItemData struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Count int    `json:"count"`
}

func NewWorldMessage(world *world.World) Message {
	return newMessage(
		MessageTypeWorld,
		worldData{
			SizeX:    world.GetSizeX(),
			SizeY:    world.GetSizeY(),
			Terrain:  world.GetTerrain(),
			Blockers: world.GetBlockers(),
			Walls:    world.GetWalls(),
			Quests:   serializeQuests(world.GetQuestRegistry().All()),
		},
	)
}

func serializeQuests(quests []world.Quest) []questData {
	result := make([]questData, len(quests))
	for i, quest := range quests {
		steps := make([]questStepData, len(quest.Steps))
		for stepIndex, step := range quest.Steps {
			steps[stepIndex] = questStepData{
				Id:          step.Id,
				Description: step.Description,
				Requirement: questRequirementData{
					EventId: step.Requirement.EventId,
					Count:   step.Requirement.Count,
				},
			}
		}
		result[i] = questData{
			Id:           quest.Id,
			DisplayName:  quest.DisplayName,
			Description:  quest.Description,
			StartEventId: quest.StartEventId,
			Steps:        steps,
			Rewards: questRewardsData{
				Items: serializeQuestRewardItems(quest.Rewards.Items),
			},
		}
	}
	return result
}

func serializeQuestRewardItems(items []world.QuestRewardItem) []questRewardItemData {
	result := make([]questRewardItemData, len(items))
	for i, item := range items {
		result[i] = questRewardItemData{
			Name:  item.Name,
			Type:  item.Type,
			Count: item.Count,
		}
	}
	return result
}
