package component

import "webscape/server/util"

const ComponentIdQuestLog = ComponentId("questlog")

type QuestProgress struct {
	QuestId          string
	StepId           string
	CurrentStepIndex int
	CurrentCount     int
}

type CQuestLog struct {
	active    map[string]*QuestProgress
	completed map[string]bool
}

func NewCQuestLog() *CQuestLog {
	return &CQuestLog{
		active:    map[string]*QuestProgress{},
		completed: map[string]bool{},
	}
}

func (c *CQuestLog) GetId() ComponentId {
	return ComponentIdQuestLog
}

func (c *CQuestLog) Serialize() util.Json {
	active := make(util.JArray, 0, len(c.active))
	for _, progress := range c.active {
		active = append(active, util.JObject(map[string]util.Json{
			"questId":          util.JString(progress.QuestId),
			"stepId":           util.JString(progress.StepId),
			"currentStepIndex": util.JNumber(progress.CurrentStepIndex),
			"currentCount":     util.JNumber(progress.CurrentCount),
		}))
	}

	completed := make(util.JArray, 0, len(c.completed))
	for questId := range c.completed {
		completed = append(completed, util.JString(questId))
	}

	return util.JObject(map[string]util.Json{
		"active":    active,
		"completed": completed,
	})
}

func (c *CQuestLog) IsActive(questId string) bool {
	_, ok := c.active[questId]
	return ok
}

func (c *CQuestLog) IsCompleted(questId string) bool {
	return c.completed[questId]
}

func (c *CQuestLog) StartQuest(questId string, stepId string) {
	if c.IsActive(questId) || c.IsCompleted(questId) {
		return
	}
	c.active[questId] = &QuestProgress{
		QuestId:          questId,
		StepId:           stepId,
		CurrentStepIndex: 0,
		CurrentCount:     0,
	}
}

func (c *CQuestLog) GetActiveProgress() []*QuestProgress {
	result := make([]*QuestProgress, 0, len(c.active))
	for _, progress := range c.active {
		result = append(result, progress)
	}
	return result
}

func (c *CQuestLog) SetProgress(questId string, stepIndex int, stepId string, currentCount int) {
	if c.completed[questId] {
		return
	}
	c.active[questId] = &QuestProgress{
		QuestId:          questId,
		StepId:           stepId,
		CurrentStepIndex: stepIndex,
		CurrentCount:     currentCount,
	}
}

func (c *CQuestLog) CompleteQuest(questId string) {
	delete(c.active, questId)
	c.completed[questId] = true
}
