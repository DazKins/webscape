package message

import (
	"encoding/json"
)

type Message struct {
	Metadata MessageMetadata `json:"metadata"`
	Data     interface{}     `json:"data"`
}

type messageDto struct {
	Metadata messageMetadataDto `json:"metadata"`
	Data     interface{}        `json:"data"`
}

func newMessage(typ MessageType, data interface{}) Message {
	return Message{
		Metadata: NewMessageMetadata(typ),
		Data:     data,
	}
}

func (m *Message) ToDto() messageDto {
	return messageDto{
		Metadata: m.Metadata.ToDto(),
		Data:     m.Data,
	}
}

func (m *Message) Marshal() string {
	dto := m.ToDto()
	json, err := json.Marshal(dto)
	if err != nil {
		panic(err)
	}
	return string(json)
}

func Unmarshal(data string) (Message, error) {
	var dto messageDto
	err := json.Unmarshal([]byte(data), &dto)
	if err != nil {
		return Message{}, err
	}

	metadata, err := dto.Metadata.FromDto()
	if err != nil {
		return Message{}, err
	}

	return Message{
		Metadata: metadata,
		Data:     dto.Data,
	}, nil
}
