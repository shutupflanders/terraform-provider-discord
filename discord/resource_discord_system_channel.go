package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/net/context"
)

func resourceDiscordSystemChannel() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSystemChannelCreate,
		ReadContext:   resourceSystemChannelRead,
		UpdateContext: resourceSystemChannelUpdate,
		DeleteContext: resourceSystemChannelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Description: "Manage the system channel of a Discord server.",
		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the server to manage the system channel for.",
			},
			"system_channel_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the channel that will be used as the system channel.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the server.",
			},
		},
	}
}

func resourceSystemChannelCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	serverId := d.Get("server_id").(string)

	server, err := executeWithRetry(ctx, func() (*discordgo.Guild, error) {
		return client.Guild(serverId, discordgo.WithContext(ctx))
	})
	if err != nil {
		return diag.Errorf("Failed to find server: %s", err.Error())
	}

	systemChannelId := server.SystemChannelID
	if v, ok := d.GetOk("system_channel_id"); ok {
		systemChannelId = v.(string)

	} else {
		return diag.Errorf("Failed to parse system channel id")
	}
	if _, err := executeWithRetry(ctx, func() (*discordgo.Guild, error) {
		return client.GuildEdit(serverId, &discordgo.GuildParams{
			SystemChannelID: systemChannelId,
		}, discordgo.WithContext(ctx))
	}); err != nil {
		return diag.Errorf("Failed to edit server: %s", err.Error())
	}

	d.SetId(serverId)

	return diags
}

func resourceSystemChannelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	serverId := d.Id()

	server, err := executeWithRetry(ctx, func() (*discordgo.Guild, error) {
		return client.Guild(serverId, discordgo.WithContext(ctx))
	})
	if err != nil {
		return diag.Errorf("Error fetching server: %s", err.Error())
	}

	d.Set("system_channel_id", server.SystemChannelID)

	return diags
}

func resourceSystemChannelUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	serverId := d.Get("server_id").(string)
	_, err := executeWithRetry(ctx, func() (*discordgo.Guild, error) {
		return client.Guild(serverId, discordgo.WithContext(ctx))
	})

	if err != nil {
		return diag.Errorf("Error fetching server: %s", err.Error())
	}

	if d.HasChange("system_channel_id") {
		id := d.Get("system_channel_id").(string)
		if _, err := executeWithRetry(ctx, func() (*discordgo.Guild, error) {
			return client.GuildEdit(serverId, &discordgo.GuildParams{
				SystemChannelID: id,
			}, discordgo.WithContext(ctx))
		}); err != nil {
			return diag.Errorf("Failed to edit server: %s", err.Error())
		}
	}

	return diags
}

func resourceSystemChannelDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	serverID := d.Get("server_id").(string)

	if _, err := executeWithRetry(ctx, func() (*discordgo.Guild, error) {
		return client.GuildEdit(serverID, &discordgo.GuildParams{
			SystemChannelID: "",
		}, discordgo.WithContext(ctx))
	}); err != nil {
		return diag.Errorf("Failed to edit server: %s: %s", serverID, err.Error())
	}

	return diags
}
