package discord

import (
	"fmt"
	"log"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"golang.org/x/net/context"
)

func resourceDiscordChannelPermission() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceChannelPermissionCreate,
		ReadContext:   resourceChannelPermissionRead,
		UpdateContext: resourceChannelPermissionUpdate,
		DeleteContext: resourceChannelPermissionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Description: "A resource to create a permission override for a channel.",
		Schema: map[string]*schema.Schema{
			"channel_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the channel for this override.",
			},
			"type": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				Description:  "Type of the override. Must be `role` or `user`.",
				ValidateFunc: validation.StringInSlice([]string{"role", "user"}, false),
			},
			"overwrite_id": {
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
				Description: "ID of the user or role for this override.",
			},
			"allow": {
				AtLeastOneOf: []string{"allow", "deny"},
				Optional:     true,
				Type:         schema.TypeInt,
				Description:  "Permission bits for the allowed permissions on this override. At least one of `allow` or `deny` must be set.",
			},
			"deny": {
				AtLeastOneOf: []string{"allow", "deny"},
				Optional:     true,
				Type:         schema.TypeInt,
				Description:  "Permission bits for the denied permissions on this override. At least one of `allow` or `deny` must be set.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Hash of the channel ID, override ID, and type.",
			},
		},
	}
}

func resourceChannelPermissionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)
	overwriteId := d.Get("overwrite_id").(string)
	permissionType, _ := getDiscordChannelPermissionType(d.Get("type").(string))
	err := executeWithRetryNoResult(ctx, func() error {
		return client.ChannelPermissionSet(
			channelId, overwriteId, permissionType,
			int64(d.Get("allow").(int)),
			int64(d.Get("deny").(int)), discordgo.WithContext(ctx))
	})
	if err != nil {
		return diag.Errorf("Failed to update channel permissions %s: %s", channelId, err.Error())
	}

	d.SetId(generateThreePartId(channelId, overwriteId, d.Get("type").(string)))

	return diags
}

func resourceChannelPermissionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		diags          diag.Diagnostics
		channelId      string
		overwriteId    string
		permissionType discordgo.PermissionOverwriteType
	)

	client := m.(*Context).Session

	cId, oId, pt, err := parseThreeIds(d.Id())
	if err != nil {
		log.Default().Printf("Unable to parse IDs out of the resource ID. Falling back on legacy config behavior.")
		channelId = d.Get("channel_id").(string)
		overwriteId = d.Get("overwrite_id").(string)
		permissionType, _ = getDiscordChannelPermissionType(d.Get("type").(string))
	} else {
		channelId = cId
		overwriteId = oId
		permissionType, _ = getDiscordChannelPermissionType(pt)

		d.Set("channel_id", channelId)
		d.Set("overwrite_id", overwriteId)
		d.Set("type", pt)
	}

	channel, err := executeWithRetry(ctx, func() (*discordgo.Channel, error) {
		return client.Channel(channelId, discordgo.WithContext(ctx))
	})
	if err != nil {
		return diag.Errorf("Failed to find channel %s: %s", channelId, err.Error())
	}

	for _, x := range channel.PermissionOverwrites {
		if uint(x.Type) == uint(permissionType) && x.ID == overwriteId {
			d.Set("allow", int(x.Allow))
			d.Set("deny", int(x.Deny))
			break
		}
	}

	return diags
}

func resourceChannelPermissionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)
	overwriteId := d.Get("overwrite_id").(string)
	permissionType, _ := getDiscordChannelPermissionType(d.Get("type").(string))

	err := executeWithRetryNoResult(ctx, func() error {
		return client.ChannelPermissionSet(
			channelId, overwriteId, permissionType,
			int64(d.Get("allow").(int)),
			int64(d.Get("deny").(int)), discordgo.WithContext(ctx))
	})
	if err != nil {
		return diag.Errorf("Failed to update channel permissions %s: %s", channelId, err.Error())
	}

	d.SetId(strconv.Itoa(
		Hashcode(
			fmt.Sprintf(
				"%s:%s:%s", channelId, overwriteId, d.Get("type").(string),
			),
		),
	),
	)
	return diags
}

func resourceChannelPermissionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)
	overwriteId := d.Get("overwrite_id").(string)

	err := executeWithRetryNoResult(ctx, func() error {
		return client.ChannelPermissionDelete(channelId, overwriteId, discordgo.WithContext(ctx))
	})
	if err != nil {
		return diag.Errorf("Failed to delete channel permissions %s: %s", channelId, err.Error())
	}

	return diags
}
