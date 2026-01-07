package model

import "github.com/google/uuid"

type ItemId uuid.UUID

func NewItemId() ItemId {
	return ItemId(uuid.New())
}

func (i ItemId) String() string {
	return uuid.UUID(i).String()
}

type Item struct {
	Id   ItemId
	Name string
	Type string
}

func NewItem(name string, itemType string) *Item {
	return &Item{
		Id:   NewItemId(),
		Name: name,
		Type: itemType,
	}
}

