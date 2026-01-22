package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDiscordForumPost() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceForumPostCreate,
		ReadContext:   resourceForumPostRead,
		UpdateContext: resourceForumPostUpdate,
		DeleteContext: resourceForumPostDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "A resource to create a forum post (thread) in a forum channel.",

		Schema: map[string]*schema.Schema{
			"channel_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The forum channel ID to create the post in.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name/title of the forum post.",
			},
			"message": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The initial message content of the post.",
			},
			"auto_archive_duration": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      10080,
				Description:  "Duration in minutes to auto-archive the thread (60, 1440, 4320, 10080).",
				ValidateFunc: validation.IntInSlice([]int{60, 1440, 4320, 10080}),
			},
			"applied_tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of tag IDs to apply to the post.",
			},
			"pinned": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the post is pinned in the forum.",
			},
			// Computed attributes
			"thread_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the created thread (same as resource ID).",
			},
			"owner_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the user who created the post.",
			},
		},
	}
}

func resourceForumPostCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)
	name := d.Get("name").(string)
	message := d.Get("message").(string)
	autoArchive := d.Get("auto_archive_duration").(int)

	// Build applied tags
	var appliedTags []string
	if v, ok := d.GetOk("applied_tags"); ok {
		for _, tag := range v.([]interface{}) {
			appliedTags = append(appliedTags, tag.(string))
		}
	}

	// Create forum post with retry handling
	thread, err := executeWithRetry(ctx, func() (*discordgo.Channel, error) {
		return client.ForumThreadStartComplex(channelId, &discordgo.ThreadStart{
			Name:                name,
			AutoArchiveDuration: autoArchive,
			AppliedTags:         appliedTags,
		}, &discordgo.MessageSend{
			Content: message,
		}, discordgo.WithContext(ctx))
	})

	if err != nil {
		return diag.Errorf("Failed to create forum post: %s", err.Error())
	}

	d.SetId(thread.ID)
	d.Set("thread_id", thread.ID)
	d.Set("owner_id", thread.OwnerID)

	// Handle pinning if requested
	if d.Get("pinned").(bool) {
		flags := discordgo.ChannelFlagPinned
		err := executeWithRetryNoResult(ctx, func() error {
			_, err := client.ChannelEditComplex(thread.ID, &discordgo.ChannelEdit{
				Flags: &flags,
			}, discordgo.WithContext(ctx))
			return err
		})
		if err != nil {
			return diag.Errorf("Failed to pin forum post: %s", err.Error())
		}
	}

	return resourceForumPostRead(ctx, d, m)
}

func resourceForumPostRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Context).Session
	var diags diag.Diagnostics

	thread, err := executeWithRetry(ctx, func() (*discordgo.Channel, error) {
		return client.Channel(d.Id(), discordgo.WithContext(ctx))
	})

	if err != nil {
		// Check if the thread was deleted (404)
		if restErr, ok := err.(*discordgo.RESTError); ok {
			if restErr.Response != nil && restErr.Response.StatusCode == 404 {
				d.SetId("")
				return diags
			}
		}
		return diag.Errorf("Failed to fetch forum post %s: %s", d.Id(), err.Error())
	}

	d.Set("channel_id", thread.ParentID)
	d.Set("name", thread.Name)
	if thread.ThreadMetadata != nil {
		d.Set("auto_archive_duration", thread.ThreadMetadata.AutoArchiveDuration)
	}
	d.Set("thread_id", thread.ID)
	d.Set("owner_id", thread.OwnerID)

	// Check if pinned
	d.Set("pinned", thread.Flags&discordgo.ChannelFlagPinned != 0)

	// Get applied tags
	if len(thread.AppliedTags) > 0 {
		d.Set("applied_tags", thread.AppliedTags)
	}

	return diags
}

func resourceForumPostUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Context).Session

	edit := &discordgo.ChannelEdit{}
	hasChanges := false

	if d.HasChange("name") {
		name := d.Get("name").(string)
		edit.Name = name
		hasChanges = true
	}

	if d.HasChange("auto_archive_duration") {
		duration := d.Get("auto_archive_duration").(int)
		edit.AutoArchiveDuration = duration
		hasChanges = true
	}

	if d.HasChange("applied_tags") {
		var tags []string
		if v, ok := d.GetOk("applied_tags"); ok {
			for _, tag := range v.([]interface{}) {
				tags = append(tags, tag.(string))
			}
		}
		edit.AppliedTags = &tags
		hasChanges = true
	}

	if d.HasChange("pinned") {
		var flags discordgo.ChannelFlags
		if d.Get("pinned").(bool) {
			flags = discordgo.ChannelFlagPinned
		} else {
			flags = 0
		}
		edit.Flags = &flags
		hasChanges = true
	}

	if hasChanges {
		err := executeWithRetryNoResult(ctx, func() error {
			_, err := client.ChannelEditComplex(d.Id(), edit, discordgo.WithContext(ctx))
			return err
		})
		if err != nil {
			return diag.Errorf("Failed to update forum post: %s", err.Error())
		}
	}

	return resourceForumPostRead(ctx, d, m)
}

func resourceForumPostDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Context).Session
	var diags diag.Diagnostics

	_, err := executeWithRetry(ctx, func() (*discordgo.Channel, error) {
		return client.ChannelDelete(d.Id(), discordgo.WithContext(ctx))
	})

	if err != nil {
		// Ignore 404 errors (already deleted)
		if restErr, ok := err.(*discordgo.RESTError); ok {
			if restErr.Response != nil && restErr.Response.StatusCode == 404 {
				return diags
			}
		}
		return diag.Errorf("Failed to delete forum post: %s", err.Error())
	}

	return diags
}
