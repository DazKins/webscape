package component

import "webscape/server/util"

const ComponentIdQuestLog = ComponentId("questlog")

type QuestProgress struct {
	QuestId          string
	StepId           string
	CurrentStepIndex int
	CurrentCount     int
}

type CompletedQuest struct {
	QuestId string
}

type CQuestLog struct {
	active       map[string]*QuestProgress
	completed    []CompletedQuest
	completedSet map[string]bool
}

func NewCQuestLog() *CQuestLog {
	return &CQuestLog{
		active:       map[string]*QuestProgress{},
		completed:    []CompletedQuest{},
		completedSet: map[string]bool{},
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
	for _, record := range c.completed {
		completed = append(completed, util.JObject(map[string]util.Json{
			"questId": util.JString(record.QuestId),
		}))
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
	return c.completedSet[questId]
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

func (c *CQuestLog) GetCompletedQuests() []CompletedQuest {
	result := make([]CompletedQuest, len(c.completed))
	copy(result, c.completed)
	return result
}

func (c *CQuestLog) SetProgress(questId string, stepIndex int, stepId string, currentCount int) {
	if c.completedSet[questId] {
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
	if c.completedSet[questId] {
		return
	}
	delete(c.active, questId)
	c.completed = append(c.completed, CompletedQuest{QuestId: questId})
	c.completedSet[questId] = true
}
