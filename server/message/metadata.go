package message

import (
	"fmt"
	"time"
)

type MessageMetadata struct {
	Type      MessageType
	Timestamp time.Time
}

type messageMetadataDto struct {
	Type      MessageTypeDto `json:"type"`
	Timestamp string         `json:"time"`
}

func NewMessageMetadata(typ MessageType) MessageMetadata {
	return MessageMetadata{
		Type:      typ,
		Timestamp: time.Now().UTC(),
	}
}

func (m *MessageMetadata) ToDto() messageMetadataDto {
	return messageMetadataDto{
		Type:      m.Type.ToDto(),
		Timestamp: m.Timestamp.Format(time.RFC3339),
	}
}

func (m *messageMetadataDto) FromDto() (MessageMetadata, error) {
	timestamp, err := time.Parse(time.RFC3339, m.Timestamp)
	if err != nil {
		return MessageMetadata{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	typ, err := m.Type.FromDto()
	if err != nil {
		return MessageMetadata{}, fmt.Errorf("failed to parse type: %w", err)
	}

	return MessageMetadata{
		Type:      typ,
		Timestamp: timestamp,
	}, nil
}
