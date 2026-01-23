package discord

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceDiscordForumPost_basic(t *testing.T) {
	testServerID := os.Getenv("DISCORD_TEST_SERVER_ID")
	if testServerID == "" {
		t.Skip("DISCORD_TEST_SERVER_ID envvar must be set for acceptance tests")
	}
	name := "discord_forum_post.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDiscordForumPost_basic(testServerID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "Test Forum Post"),
					resource.TestCheckResourceAttrSet(name, "thread_id"),
					resource.TestCheckResourceAttrSet(name, "channel_id"),
					resource.TestCheckResourceAttr(name, "auto_archive_duration", "10080"),
					resource.TestCheckResourceAttr(name, "pinned", "false"),
				),
			},
		},
	})
}

func TestAccResourceDiscordForumPost_pinned(t *testing.T) {
	testServerID := os.Getenv("DISCORD_TEST_SERVER_ID")
	if testServerID == "" {
		t.Skip("DISCORD_TEST_SERVER_ID envvar must be set for acceptance tests")
	}
	name := "discord_forum_post.pinned"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDiscordForumPost_pinned(testServerID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "Pinned Forum Post"),
					resource.TestCheckResourceAttr(name, "pinned", "true"),
					resource.TestCheckResourceAttrSet(name, "thread_id"),
				),
			},
		},
	})
}

func TestAccResourceDiscordForumPost_update(t *testing.T) {
	testServerID := os.Getenv("DISCORD_TEST_SERVER_ID")
	if testServerID == "" {
		t.Skip("DISCORD_TEST_SERVER_ID envvar must be set for acceptance tests")
	}
	name := "discord_forum_post.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDiscordForumPost_basic(testServerID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "Test Forum Post"),
				),
			},
			{
				Config: testAccResourceDiscordForumPost_updated(testServerID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "Updated Forum Post"),
					resource.TestCheckResourceAttr(name, "auto_archive_duration", "4320"),
				),
			},
		},
	})
}

func testAccResourceDiscordForumPost_basic(serverID string) string {
	return fmt.Sprintf(`
resource "discord_forum_channel" "test" {
  server_id = "%[1]s"
  name      = "terraform-test-forum"
}

resource "discord_forum_post" "test" {
  channel_id            = discord_forum_channel.test.id
  name                  = "Test Forum Post"
  message               = "This is a test forum post created by Terraform"
  auto_archive_duration = 10080
}`, serverID)
}

func testAccResourceDiscordForumPost_pinned(serverID string) string {
	return fmt.Sprintf(`
resource "discord_forum_channel" "test_pinned" {
  server_id = "%[1]s"
  name      = "terraform-test-forum-pinned"
}

resource "discord_forum_post" "pinned" {
  channel_id = discord_forum_channel.test_pinned.id
  name       = "Pinned Forum Post"
  message    = "This is a pinned forum post"
  pinned     = true
}`, serverID)
}

func testAccResourceDiscordForumPost_updated(serverID string) string {
	return fmt.Sprintf(`
resource "discord_forum_channel" "test" {
  server_id = "%[1]s"
  name      = "terraform-test-forum"
}

resource "discord_forum_post" "test" {
  channel_id            = discord_forum_channel.test.id
  name                  = "Updated Forum Post"
  message               = "This is a test forum post created by Terraform"
  auto_archive_duration = 4320
}`, serverID)
}
