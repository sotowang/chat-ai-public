package repo

import (
	"chat-ai/chat-server/model"
)

func QueryPromptByRole(role string) (string, error) {
	var rolePrompt model.RolePrompt
	err := MyDB.Table("role_prompt").Where("role = ?", role).First(&rolePrompt).Error
	if err != nil {
		return "", err
	} else {
		return rolePrompt.Prompt, nil
	}
}

func QueryAllRolePrompt() ([]*model.RolePrompt, error) {
	var rolePrompts []*model.RolePrompt
	err := MyDB.Table("role_prompt").Find(&rolePrompts).Error
	if err != nil {
		return nil, err
	}
	return rolePrompts, nil
}

func QueryAllRoleByUserId(userId uint) ([]*model.RolePrompt, error) {
	var rolePrompts []*model.RolePrompt
	err := MyDB.Table("role_prompt").
		Where("user_id = ? OR user_id = 0", userId).
		Find(&rolePrompts).Error
	if err != nil {
		return nil, err
	}
	return rolePrompts, nil
}
