package world

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
)

type QuestRegistry struct {
	quests map[string]*Quest
}

type questDocument struct {
	FormatVersion int     `json:"formatVersion"`
	Id            string  `json:"id"`
	DisplayName   string  `json:"displayName"`
	Quests        []Quest `json:"quests"`
}

type Quest struct {
	Id           string      `json:"id"`
	DisplayName  string      `json:"displayName"`
	Description  string      `json:"description"`
	StartEventId string      `json:"startEventId"`
	Steps        []QuestStep `json:"steps"`
}

type QuestStep struct {
	Id          string           `json:"id"`
	Description string           `json:"description"`
	Requirement QuestRequirement `json:"requirement"`
}

type QuestRequirement struct {
	EventId string `json:"eventId"`
	Count   int    `json:"count"`
}

func NewQuestRegistry() *QuestRegistry {
	return &QuestRegistry{
		quests: map[string]*Quest{},
	}
}

func loadQuestRegistry(gameFS fs.FS, paths []string) (*QuestRegistry, error) {
	registry := NewQuestRegistry()
	for _, path := range paths {
		data, err := fs.ReadFile(gameFS, path)
		if err != nil {
			return nil, fmt.Errorf("read quest %q: %w", path, err)
		}

		var document questDocument
		if err := json.Unmarshal(data, &document); err != nil {
			return nil, fmt.Errorf("parse quest %q: %w", path, err)
		}
		if err := validateQuestDocument(document); err != nil {
			return nil, fmt.Errorf("validate quest %q: %w", path, err)
		}

		for i := range document.Quests {
			quest := &document.Quests[i]
			if _, exists := registry.quests[quest.Id]; exists {
				return nil, fmt.Errorf("duplicate quest id %q", quest.Id)
			}
			registry.quests[quest.Id] = quest
		}
	}

	return registry, nil
}

func (r *QuestRegistry) Get(id string) (*Quest, bool) {
	if r == nil {
		return nil, false
	}
	quest, ok := r.quests[id]
	return quest, ok
}

func (r *QuestRegistry) All() []Quest {
	if r == nil {
		return []Quest{}
	}
	quests := make([]Quest, 0, len(r.quests))
	for _, quest := range r.quests {
		quests = append(quests, *quest)
	}
	return quests
}

func (r *QuestRegistry) Len() int {
	if r == nil {
		return 0
	}
	return len(r.quests)
}

func validateQuestDocument(document questDocument) error {
	if document.FormatVersion != 1 {
		return fmt.Errorf("unsupported quest format version %d", document.FormatVersion)
	}
	if document.Id == "" {
		return errors.New("quest document id is required")
	}
	if len(document.Quests) == 0 {
		return errors.New("quest document must include at least one quest")
	}

	seenQuestIds := map[string]bool{}
	for _, quest := range document.Quests {
		if quest.Id == "" {
			return errors.New("quest id is required")
		}
		if seenQuestIds[quest.Id] {
			return fmt.Errorf("duplicate quest id %q", quest.Id)
		}
		seenQuestIds[quest.Id] = true
		if len(quest.Steps) == 0 {
			return fmt.Errorf("quest %q must include at least one step", quest.Id)
		}

		seenStepIds := map[string]bool{}
		for _, step := range quest.Steps {
			if step.Id == "" {
				return fmt.Errorf("quest %q has a step without an id", quest.Id)
			}
			if seenStepIds[step.Id] {
				return fmt.Errorf("quest %q has duplicate step id %q", quest.Id, step.Id)
			}
			seenStepIds[step.Id] = true
			if step.Requirement.EventId == "" {
				return fmt.Errorf("quest %q step %q must include requirement eventId", quest.Id, step.Id)
			}
			if step.Requirement.Count < 1 {
				return fmt.Errorf("quest %q step %q requirement count must be at least 1", quest.Id, step.Id)
			}
		}
	}

	return nil
}
