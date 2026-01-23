package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/net/context"
)

func resourceDiscordInvite() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceInviteCreate,
		ReadContext:   resourceInviteRead,
		DeleteContext: resourceInviteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Description: "A resource to create an invite for a channel.",
		Schema: map[string]*schema.Schema{
			"channel_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the channel to create an invite for.",
			},
			"max_age": {
				Type:        schema.TypeInt,
				ForceNew:    true,
				Optional:    true,
				Default:     86400,
				Description: "Age of the invite. `0` for permanent. (default `86400`)",
			},
			"max_uses": {
				Type:        schema.TypeInt,
				ForceNew:    true,
				Optional:    true,
				Description: "Max number of uses for the invite. `0` (the default) for unlimited.",
			},
			"temporary": {
				Type:        schema.TypeBool,
				ForceNew:    true,
				Optional:    true,
				Description: "Whether the invite kicks users after they close Discord. (default `false`)",
			},
			"unique": {
				Type:        schema.TypeBool,
				ForceNew:    true,
				Optional:    true,
				Description: "Whether this should create a new invite every time.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The invite code.",
			},
			"code": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The invite code.",
			},
		},
	}
}

func resourceInviteCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)

	invite, err := executeWithRetry(ctx, func() (*discordgo.Invite, error) {
		return client.ChannelInviteCreate(channelId, discordgo.Invite{
			MaxAge:    d.Get("max_age").(int),
			MaxUses:   d.Get("max_uses").(int),
			Temporary: d.Get("temporary").(bool),
			Unique:    d.Get("unique").(bool),
		}, discordgo.WithContext(ctx))
	})
	if err != nil {
		return diag.Errorf("Failed to create a invite: %s", err.Error())
	}
	d.SetId(invite.Code)
	d.Set("code", invite.Code)

	return diags
}

func resourceInviteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	invite, err := executeWithRetry(ctx, func() (*discordgo.Invite, error) {
		return client.Invite(d.Id(), discordgo.WithContext(ctx))
	})
	if err != nil {
		d.SetId("")
	} else {
		d.Set("code", invite.Code)
	}

	return diags
}

func resourceInviteDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	if err := executeWithRetryNoResult(ctx, func() error {
		_, err := client.InviteDelete(d.Id(), discordgo.WithContext(ctx))
		return err
	}); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
