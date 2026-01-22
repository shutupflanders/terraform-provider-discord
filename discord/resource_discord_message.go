package discord

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/net/context"
)

func resourceDiscordMessage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMessageCreate,
		ReadContext:   resourceMessageRead,
		UpdateContext: resourceMessageUpdate,
		DeleteContext: resourceMessageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Description: "A resource to create a message",
		Schema: map[string]*schema.Schema{
			"channel_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the channel the message will be in.",
			},
			"server_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the server this message is in.",
			},
			"author": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the user who wrote the message.",
			},
			"content": {
				AtLeastOneOf: []string{"content", "embed"},
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Text content of message. At least one of `content` or `embed` must be set.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == strings.TrimSuffix(new, "\r\n")
				},
			},
			"timestamp": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "When the message was sent.",
			},
			"edited_timestamp": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "When the message was edited.",
			},
			"tts": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether this message triggers TTS. (default `false`)",
			},
			"embed": {
				AtLeastOneOf: []string{"content", "embed"},
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				Description:  "An embed block. At least one of `content` or `embed` must be set.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"title": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Title of the embed.",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Description of the embed.",
						},
						"url": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "URL of the embed.",
						},
						"timestamp": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Timestamp of the embed content.",
						},
						"color": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Color of the embed. Must be an integer color code.",
						},
						"footer": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Footer of the embed.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"text": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Text of the footer.",
									},
									"icon_url": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "URL to an icon to be included in the footer.",
									},
								},
							},
						},
						"image": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Image to be included in the embed.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"url": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "URL of the image to be included in the embed.",
									},
									"proxy_url": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "URL to access the image via Discord's proxy.",
									},
									"height": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Height of the image.",
									},
									"width": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Width of the image.",
									},
								},
							},
						},
						"thumbnail": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Thumbnail to be included in the embed.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"url": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "URL of the thumbnail to be included in the embed.",
									},
									"proxy_url": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "URL to access the thumbnail via Discord's proxy.",
									},
									"height": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Height of the thumbnail.",
									},
									"width": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Width of the thumbnail.",
									},
								},
							},
						},
						"video": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Video to be included in the embed.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"url": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "URL of the video to be included in the embed.",
									},
									"height": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Height of the video.",
									},
									"width": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Width of the video.",
									},
								},
							},
						},
						"provider": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Provider of the embed.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Name of the provider.",
									},
									"url": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "URL of the provider.",
									},
								},
							},
						},
						"author": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Author of the embed.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Name of the author.",
									},
									"url": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "URL of the author.",
									},
									"icon_url": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "URL of the author's icon.",
									},
									"proxy_icon_url": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "URL to access the author's icon via Discord's proxy.",
									},
								},
							},
						},
						"fields": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Fields of the embed.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Name of the field.",
									},
									"value": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Value of the field.",
									},
									"inline": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Whether the field is inline.",
									},
								},
							},
						},
					},
				},
			},
			"pinned": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether this message is pinned. (default `false`)",
			},
			"type": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The type of the message.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the message.",
			},
		},
	}
}

func resourceMessageCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)

	embeds := make([]*discordgo.MessageEmbed, 0, 1)
	if v, ok := d.GetOk("embed"); ok {
		if embed, err := buildEmbed(v.([]interface{})); err != nil {
			return diag.Errorf("Failed to create message in %s: %s", channelId, err.Error())
		} else {
			embeds = append(embeds, embed)
		}
	}
	message, err := executeWithRetry(ctx, func() (*discordgo.Message, error) {
		return client.ChannelMessageSendComplex(channelId, &discordgo.MessageSend{
			Content: d.Get("content").(string),
			Embeds:  embeds,
			TTS:     d.Get("tts").(bool),
		}, discordgo.WithContext(ctx))
	})
	if err != nil {
		return diag.Errorf("Failed to create message in %s: %s", channelId, err.Error())
	}

	d.SetId(message.ID)
	d.Set("type", int(message.Type))
	d.Set("timestamp", message.Timestamp.Format(time.RFC3339))
	d.Set("author", message.Author.ID)
	if len(message.Embeds) > 0 {
		d.Set("embed", unbuildEmbed(message.Embeds[0]))
	} else {
		d.Set("embed", nil)
	}
	if message.GuildID != "" {
		d.Set("server_id", message.GuildID)
	}

	if d.Get("pinned").(bool) {
		pinError := executeWithRetryNoResult(ctx, func() error {
			return client.ChannelMessagePin(channelId, message.ID, discordgo.WithContext(ctx))
		})
		if pinError != nil {
			diags = append(diags, diag.Errorf("Failed to pin message %s in %s: %s", message.ID, channelId, pinError.Error())...)
		}
	}

	return diags
}

func resourceMessageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)
	messageId := d.Id()
	message, err := executeWithRetry(ctx, func() (*discordgo.Message, error) {
		return client.ChannelMessage(channelId, messageId, discordgo.WithContext(ctx))
	})
	if err != nil {
		return diag.Errorf("Failed to fetch message %s in %s: %s", messageId, channelId, err.Error())
	}

	if message.GuildID != "" {
		d.Set("server_id", message.GuildID)
	}
	d.Set("type", int(message.Type))
	d.Set("tts", message.TTS)
	d.Set("timestamp", message.Timestamp.Format(time.RFC3339))
	d.Set("author", message.Author.ID)
	d.Set("content", message.Content)
	d.Set("pinned", message.Pinned)

	if len(message.Embeds) > 0 {
		d.Set("embed", unbuildEmbed(message.Embeds[0]))
	}
	if message.EditedTimestamp != nil {
		d.Set("edited_timestamp", message.EditedTimestamp.Format(time.RFC3339))
	}

	return diags
}

func resourceMessageUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)
	messageId := d.Id()

	var content string
	message, err := executeWithRetry(ctx, func() (*discordgo.Message, error) {
		return client.ChannelMessage(channelId, messageId, discordgo.WithContext(ctx))
	})
	if err != nil {
		return diag.Errorf("Failed to fetch message %s in %s: %s", messageId, channelId, err.Error())
	}
	if d.HasChange("content") {
		content = d.Get("content").(string)
	} else {
		content = message.Content
	}

	embeds := make([]*discordgo.MessageEmbed, 0, 1)
	if d.HasChange("embed") {
		var embed *discordgo.MessageEmbed
		_, n := d.GetChange("embed")
		if len(n.([]interface{})) > 0 {
			if e, err := buildEmbed(n.([]interface{})); err != nil {
				return diag.Errorf("Failed to edit message %s in %s: %s", messageId, channelId, err.Error())
			} else {
				embed = e
			}
		}

		embeds = append(embeds, embed)
	}

	editedMessage, err := executeWithRetry(ctx, func() (*discordgo.Message, error) {
		return client.ChannelMessageEditComplex(&discordgo.MessageEdit{
			ID:      messageId,
			Channel: channelId,
			Content: &content,
			Embeds:  &embeds,
		}, discordgo.WithContext(ctx))
	})
	if err != nil {
		return diag.Errorf("Failed to update message %s in %s: %s", channelId, messageId, err.Error())
	}

	if len(editedMessage.Embeds) > 0 {
		d.Set("embed", unbuildEmbed(message.Embeds[0]))
	} else {
		d.Set("embed", nil)
	}

	d.Set("edited_timestamp", editedMessage.EditedTimestamp.Format(time.RFC3339))

	return diags
}

func resourceMessageDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)
	messageId := d.Id()
	err := executeWithRetryNoResult(ctx, func() error {
		return client.ChannelMessageDelete(channelId, messageId, discordgo.WithContext(ctx))
	})
	if err != nil {
		return diag.Errorf("Failed to delete message %s in %s: %s", messageId, channelId, err.Error())
	}

	return diags
}
