package component

import (
	"fmt"
)

type savedQuestLog struct {
	Active    map[string]*QuestProgress `json:"active"`
	Completed []CompletedQuest          `json:"completed"`
}

func (c *CQuestLog) Save() (SavedComponent, error) {
	return marshalSaved(savedQuestLog{Active: c.active, Completed: c.completed})
}
func init() { registerComponentRestore(ComponentIdQuestLog, 1, restoreQuestLog) }
func restoreQuestLog(saved SavedComponent) (Component, error) {
	var s savedQuestLog
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	c := NewCQuestLog()
	for id, p := range s.Active {
		if p == nil || p.QuestId != id || p.CurrentStepIndex < 0 || p.CurrentCount < 0 {
			return nil, fmt.Errorf("invalid quest progress")
		}
		c.SetProgress(id, p.CurrentStepIndex, p.StepId, p.CurrentCount)
	}
	for _, q := range s.Completed {
		c.CompleteQuest(q.QuestId)
	}
	return c, nil
}

func (c *CQuestLog) ValidateSaved(ctx SaveContext) error {
	for _, p := range c.GetActiveProgress() {
		if !ctx.ValidQuestStep(p.QuestId, p.CurrentStepIndex, p.StepId) {
			return fmt.Errorf("invalid saved quest step")
		}
	}
	for _, q := range c.GetCompletedQuests() {
		if !ctx.HasQuest(q.QuestId) {
			return fmt.Errorf("unknown completed quest")
		}
	}

	return nil
}
