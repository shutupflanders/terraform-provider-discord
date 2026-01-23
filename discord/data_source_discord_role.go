package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDiscordRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDiscordRoleRead,
		Description: "Fetches a role's information from a server.",
		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The server ID to search for the user in.",
			},
			"role_id": {
				ExactlyOneOf: []string{"role_id", "name"},
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The user ID to search for. Either this or `name` is required.",
			},
			"name": {
				ExactlyOneOf: []string{"role_id", "name"},
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The role name to search for. Either this or `role_id` is required.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the role.",
			},
			"position": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Position of the role. This is reverse-indexed, with `@everyone` being `0`.",
			},
			"color": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The integer representation of the role's color with decimal color code.",
			},
			"permissions": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The permission bits of the role.",
			},
			"hoist": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the role is hoisted.",
			},
			"mentionable": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the role is mentionable.",
			},
			"managed": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the role is managed.",
			},
		},
	}
}

func dataSourceDiscordRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error
	var role *discordgo.Role
	client := m.(*Context).Session

	serverId := d.Get("server_id").(string)
	server, err := executeWithRetry(ctx, func() (*discordgo.Guild, error) {
		return client.Guild(serverId, discordgo.WithContext(ctx))
	})
	if err != nil {
		return diag.Errorf("Failed to fetch server %s: %s", serverId, err.Error())
	}

	roleID := d.Get("role_id").(string)
	roleName := d.Get("name").(string)
	for _, r := range server.Roles {
		if r.ID == roleID || r.Name == roleName {
			role = r
			break
		}
	}
	if role == nil {
		return diag.Errorf("Failed to find role by ID %s or name: %s", roleID, roleName)
	}

	d.SetId(role.ID)
	d.Set("role_id", role.ID)
	d.Set("name", role.Name)
	d.Set("position", role.Position)
	d.Set("color", role.Color)
	d.Set("hoist", role.Hoist)
	d.Set("mentionable", role.Mentionable)
	d.Set("permissions", int(role.Permissions))
	d.Set("managed", role.Managed)

	return diags
}
