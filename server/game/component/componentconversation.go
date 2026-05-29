package component

import "webscape/server/game/model"

const ComponentIdConversation = ComponentId("conversation")
const ComponentIdActiveConversation = ComponentId("activeconversation")

type CConversation struct {
	conversationId string
}

func NewCConversation(conversationId string) *CConversation {
	return &CConversation{
		conversationId: conversationId,
	}
}

func (c *CConversation) GetId() ComponentId {
	return ComponentIdConversation
}

func (c *CConversation) GetConversationId() string {
	return c.conversationId
}

type CActiveConversation struct {
	conversationId string
	targetEntityId model.EntityId
	currentNodeId  string
}

func NewCActiveConversation(
	conversationId string,
	targetEntityId model.EntityId,
	currentNodeId string,
) *CActiveConversation {
	return &CActiveConversation{
		conversationId: conversationId,
		targetEntityId: targetEntityId,
		currentNodeId:  currentNodeId,
	}
}

func (c *CActiveConversation) GetId() ComponentId {
	return ComponentIdActiveConversation
}

func (c *CActiveConversation) GetConversationId() string {
	return c.conversationId
}

func (c *CActiveConversation) GetTargetEntityId() model.EntityId {
	return c.targetEntityId
}

func (c *CActiveConversation) GetCurrentNodeId() string {
	return c.currentNodeId
}

func (c *CActiveConversation) SetCurrentNodeId(currentNodeId string) {
	c.currentNodeId = currentNodeId
}
