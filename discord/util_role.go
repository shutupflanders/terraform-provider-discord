package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

type Role struct {
	ServerId string
	RoleId   string
	Role     *discordgo.Role
}

func insertRole(array []*discordgo.Role, value *discordgo.Role, index int) []*discordgo.Role {
	return append(array[:index], append([]*discordgo.Role{value}, array[index:]...)...)
}

func removeRole(array []*discordgo.Role, index int) []*discordgo.Role {
	return append(array[:index], array[index+1:]...)
}

func removeRoleById(array []string, id string) []string {
	roles := make([]string, 0, len(array))
	for _, x := range array {
		if x != id {
			roles = append(roles, x)
		}
	}

	return roles
}

func moveRole(array []*discordgo.Role, srcIndex int, dstIndex int) []*discordgo.Role {
	value := array[srcIndex]
	return insertRole(removeRole(array, srcIndex), value, dstIndex)
}

func findRoleIndex(array []*discordgo.Role, value *discordgo.Role) (int, bool) {
	for index, element := range array {
		if element.ID == value.ID {
			return index, true
		}
	}

	return -1, false
}

func findRoleById(array []*discordgo.Role, id string) *discordgo.Role {
	for _, element := range array {
		if element.ID == id {
			return element
		}
	}

	return nil
}

func reorderRoles(ctx context.Context, m interface{}, serverId string, role *discordgo.Role, position int) (bool, diag.Diagnostics) {
	client := m.(*Context).Session

	roles, err := executeWithRetry(ctx, func() ([]*discordgo.Role, error) {
		return client.GuildRoles(serverId, discordgo.WithContext(ctx))
	})
	if err != nil {
		return false, diag.Errorf("Failed to fetch roles: %s", err.Error())
	}
	index, exists := findRoleIndex(roles, role)
	if !exists {
		return false, diag.Errorf("Role somehow does not exists")
	}

	moveRole(roles, index, position)

	if roles, err = executeWithRetry(ctx, func() ([]*discordgo.Role, error) {
		return client.GuildRoleReorder(serverId, roles, discordgo.WithContext(ctx))
	}); err != nil {
		return false, diag.Errorf("Failed to re-order roles: %s", err.Error())
	}

	return true, nil
}

func getRole(ctx context.Context, client *discordgo.Session, serverId string, roleId string) (*discordgo.Role, error) {
	roles, err := executeWithRetry(ctx, func() ([]*discordgo.Role, error) {
		return client.GuildRoles(serverId, discordgo.WithContext(ctx))
	})
	if err != nil {
		return nil, err
	}
	return findRoleById(roles, roleId), nil
}
