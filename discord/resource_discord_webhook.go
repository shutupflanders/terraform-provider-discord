package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/polds/imgbase64"
	"golang.org/x/net/context"
)

func resourceDiscordWebhook() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWebhookCreate,
		ReadContext:   resourceWebhookRead,
		UpdateContext: resourceWebhookUpdate,
		DeleteContext: resourceWebhookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Description: "A resource to create a webhook for a channel.",
		Schema: map[string]*schema.Schema{
			"channel_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the channel to create a webhook for.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Default name of the webhook.",
			},
			"avatar_url": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"avatar_data_uri"},
				Description:   "Remote URL for setting the default avatar of the webhook.",
			},
			"avatar_data_uri": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"avatar_url"},
				Description:   "Data URI of an image to set as the default avatar of the webhook.",
			},
			"avatar_hash": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Hash of the avatar.",
			},
			"token": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The webhook token.",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The webhook URL.",
			},
			"slack_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The Slack-compatible webhook URL.",
			},
			"github_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The GitHub-compatible webhook URL.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the webhook.",
			},
		},
	}
}

func resourceWebhookCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)

	avatar := ""
	if v, ok := d.GetOk("avatar_url"); ok {
		avatar = imgbase64.FromRemote(v.(string))
	}
	if v, ok := d.GetOk("avatar_data_uri"); ok {
		avatar = v.(string)
	}

	webhook, err := executeWithRetry(ctx, func() (*discordgo.Webhook, error) {
		return client.WebhookCreate(channelId, d.Get("name").(string), avatar, discordgo.WithContext(ctx))
	})
	if err != nil {
		return diag.Errorf("Failed to create webhook: %s", err.Error())
	}

	url := "https://discord.com/api/webhooks/" + webhook.ID + "/" + webhook.Token

	d.SetId(webhook.ID)
	d.Set("avatar_hash", webhook.Avatar)
	d.Set("token", webhook.Token)
	d.Set("url", url)
	d.Set("slack_url", url+"/slack")
	d.Set("github_url", url+"/github")

	return diags
}

func resourceWebhookRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	webhook, err := executeWithRetry(ctx, func() (*discordgo.Webhook, error) {
		return client.Webhook(d.Id(), discordgo.WithContext(ctx))
	})
	if err != nil {
		d.SetId("")
		return diags
	}

	url := "https://discord.com/api/webhooks/" + webhook.ID + "/" + webhook.Token

	d.Set("channel_id", webhook.ChannelID)
	d.Set("name", webhook.Name)
	d.Set("avatar_hash", webhook.Avatar)
	d.Set("token", webhook.Token)
	d.Set("url", url)
	d.Set("slack_url", url+"/slack")
	d.Set("github_url", url+"/github")

	return diags
}

func resourceWebhookUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)
	name := d.Get("name").(string)

	avatar := ""
	if v, ok := d.GetOk("avatar_url"); ok {
		avatar = imgbase64.FromRemote(v.(string))
	}
	if v, ok := d.GetOk("avatar_data_uri"); ok {
		avatar = v.(string)
	}

	webhook, err := executeWithRetry(ctx, func() (*discordgo.Webhook, error) {
		return client.WebhookEdit(d.Id(), name, avatar, channelId, discordgo.WithContext(ctx))
	})
	if err != nil {
		return diag.Errorf("Failed to update webhook %s: %s", d.Id(), err.Error())
	}

	url := "https://discord.com/api/webhooks/" + webhook.ID + "/" + webhook.Token
	d.Set("channel_id", webhook.ChannelID)
	d.Set("name", webhook.Name)
	d.Set("avatar_hash", webhook.Avatar)
	d.Set("token", webhook.Token)
	d.Set("url", url)
	d.Set("slack_url", url+"/slack")
	d.Set("github_url", url+"/github")

	return diags
}

func resourceWebhookDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	err := executeWithRetryNoResult(ctx, func() error {
		return client.WebhookDelete(d.Id(), discordgo.WithContext(ctx))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
