package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
	"github.com/slack-go/slack"
	"k8s.io/klog/v2"
)

type Slack struct {
	WebhookUrl     string
	DefaultChannel string // Slack channel name
	Username       string // Slack username (will show in slack message)
	ClusterName    string // Kubernete cluster name (will show in slack message)
	MuteSeconds    int    // The time to mute duplicate alerts
	// History stores sent alerts, key: Namespace/podName, value: sentTime
	History map[string]time.Time
}

type SlackMessage struct {
	Title  string
	Text   string
	Footer string
}

// Notifier is a generic interface for sending notifications.
type Notifier interface {
	SendToChannel(msg SlackMessage, channel string) error
	GetHistory() map[string]time.Time
	GetMuteSeconds() int
	SetHistory(key string, t time.Time)
	GetMaxMessageLength() int
}

// DiscordNotifier implements Notifier for Discord webhooks.
type DiscordNotifier struct {
	WebhookUrl     string
	DefaultChannel string // Discord channel ID
	Username       string // Bot username
	ClusterName    string
	MuteSeconds    int
	History        map[string]time.Time
}

func NewSlack() *Slack {
	var slackWebhookUrl, slackChannel, slackUsername, clusterName string

	if slackWebhookUrl = os.Getenv("SLACK_WEBHOOK_URL"); slackWebhookUrl == "" {
		klog.Exit("Environment variable SLACK_WEBHOOK_URL is not set")
	}

	if slackChannel = os.Getenv("SLACK_CHANNEL"); slackChannel == "" {
		slackChannel = "restart-info-nonprod"
		klog.Warningf("Environment variable SLACK_CHANNEL is not set, default: %s\n", slackChannel)
	}

	if slackUsername = os.Getenv("SLACK_USERNAME"); slackUsername == "" {
		slackUsername = "k8s-pod-restart-info-collector"
		klog.Warningf("Environment variable SLACK_USERNAME is not set, default: %s\n", slackUsername)
	}

	if clusterName = os.Getenv("CLUSTER_NAME"); clusterName == "" {
		clusterName = "cluster-name"
		klog.Warningf("Environment variable CLUSTER_NAME is not set, default: %s\n", clusterName)
	}

	muteSeconds, err := strconv.Atoi(os.Getenv("MUTE_SECONDS"))
	if err != nil {
		muteSeconds = 600
		klog.Warningf("Environment variable MUTE_SECONDS is not set, default: %d\n", muteSeconds)
	}

	klog.Infof("Slack Info: channel: %s, username: %s, clustername: %s, muteseconds: %d\n", slackChannel, slackUsername, clusterName, muteSeconds)

	return &Slack{
		WebhookUrl:     slackWebhookUrl,
		DefaultChannel: slackChannel,
		Username:       slackUsername,
		ClusterName:    clusterName,
		MuteSeconds:    muteSeconds,
		History:        make(map[string]time.Time),
	}
}

func (s *Slack) SendToChannel(msg SlackMessage, slackChannel string) error {
	channel := s.DefaultChannel
	if slackChannel != "" {
		channel = slackChannel
	}

	attachment := slack.Attachment{
		Text:       msg.Text,
		Pretext:    msg.Title,
		Footer:     msg.Footer,
		MarkdownIn: []string{"text", "pretext"},
		Color:      "#4599DF",
		Ts:         json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
	}

	err := slack.PostWebhook(s.WebhookUrl, &slack.WebhookMessage{
		Username:    s.Username,
		Channel:     channel,
		IconEmoji:   ":kubernetes:",
		Attachments: []slack.Attachment{attachment},
	})
	if err != nil {
		klog.Errorf("Sending to Slack channel failed with %v", err)
		return err
	}
	klog.Infof("Sent: [%s] to Slack.\n\n", strings.Replace(msg.Title, "\n", " ", -1))
	return nil
}

func (s *Slack) GetHistory() map[string]time.Time {
	return s.History
}
func (s *Slack) GetMuteSeconds() int {
	return s.MuteSeconds
}
func (s *Slack) SetHistory(key string, t time.Time) {
	s.History[key] = t
}

func (s *Slack) GetMaxMessageLength() int {
	return 7500
}

func NewDiscordNotifier() *DiscordNotifier {
	var discordWebhookUrl, discordChannel, discordUsername, clusterName string

	if discordWebhookUrl = os.Getenv("DISCORD_WEBHOOK_URL"); discordWebhookUrl == "" {
		klog.Exit("Environment variable DISCORD_WEBHOOK_URL is not set")
	}
	if discordChannel = os.Getenv("DISCORD_CHANNEL_ID"); discordChannel == "" {
		discordChannel = "" // Discord webhook can be channel-specific
		klog.Warningf("Environment variable DISCORD_CHANNEL_ID is not set, using webhook default")
	}
	if discordUsername = os.Getenv("DISCORD_USERNAME"); discordUsername == "" {
		discordUsername = "k8s-pod-restart-info-collector"
		klog.Warningf("Environment variable DISCORD_USERNAME is not set, default: %s\n", discordUsername)
	}
	if clusterName = os.Getenv("CLUSTER_NAME"); clusterName == "" {
		clusterName = "cluster-name"
		klog.Warningf("Environment variable CLUSTER_NAME is not set, default: %s\n", clusterName)
	}
	muteSeconds, err := strconv.Atoi(os.Getenv("MUTE_SECONDS"))
	if err != nil {
		muteSeconds = 600
		klog.Warningf("Environment variable MUTE_SECONDS is not set, default: %d\n", muteSeconds)
	}
	klog.Infof("Discord Info: channel: %s, username: %s, clustername: %s, muteseconds: %d\n", discordChannel, discordUsername, clusterName, muteSeconds)
	return &DiscordNotifier{
		WebhookUrl:     discordWebhookUrl,
		DefaultChannel: discordChannel,
		Username:       discordUsername,
		ClusterName:    clusterName,
		MuteSeconds:    muteSeconds,
		History:        make(map[string]time.Time),
	}
}

func (d *DiscordNotifier) SendToChannel(msg SlackMessage, channel string) error {
	webhookUrl := d.WebhookUrl
	content := "**" + msg.Title + "**\n" + msg.Text + "\n" + msg.Footer
	klog.Infof("Content length: %d", utf8.RuneCountInString(content))
	params := &discordgo.WebhookParams{
		Content:  content,
		Username: d.Username,
	}
	// Parse webhook URL: https://discord.com/api/webhooks/{webhook.id}/{webhook.token}
	re := regexp.MustCompile(`https://discord.com/api/webhooks/([^/]+)/([^/]+)`)
	matches := re.FindStringSubmatch(webhookUrl)
	if len(matches) != 3 {
		return fmt.Errorf("invalid Discord webhook URL format")
	}
	webhookID := matches[1]
	webhookToken := matches[2]
	dg, err := discordgo.New("")
	if err != nil {
		return err
	}
	_, err = dg.WebhookExecute(webhookID, webhookToken, false, params)
	if err != nil {
		klog.Errorf("Sending to Discord channel failed with %v", err)
		return err
	}
	klog.Infof("Sent: [%s] to Discord.\n\n", strings.Replace(msg.Title, "\n", " ", -1))
	return nil
}

func (d *DiscordNotifier) GetHistory() map[string]time.Time {
	return d.History
}
func (d *DiscordNotifier) GetMuteSeconds() int {
	return d.MuteSeconds
}
func (d *DiscordNotifier) SetHistory(key string, t time.Time) {
	d.History[key] = t
}

func (d *DiscordNotifier) GetMaxMessageLength() int {
	return 1600
}
