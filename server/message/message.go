package message

import (
	"encoding/json"
)

type Message struct {
	Metadata MessageMetadata `json:"metadata"`
	Data     any             `json:"data"`
}

type messageDto struct {
	Metadata messageMetadataDto `json:"metadata"`
	Data     any                `json:"data"`
}

func newMessage(typ MessageType, data any) Message {
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
