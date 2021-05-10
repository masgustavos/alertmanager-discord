package discord

// WebhookParams defines the message body expected by Discord's API
type WebhookParams struct {
	Content   string         `json:"content,omitempty"`
	Username  string         `json:"username,omitempty"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Embeds    []MessageEmbed `json:"embeds,omitempty"`
}

// MessageEmbed contains some of the available fields in Discord Embeds
type MessageEmbed struct {
	URL         string `json:"url,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
	Color       int    `json:"color,omitempty"`
}

type EmbedQueueItem struct {
	Embed    MessageEmbed `json:"messageEmbed,omitempty"`
	Priority int          `json:"priority"`
}

// A EmbedQueue holds embeds to be ordered before being sent to discord.
type EmbedQueue []EmbedQueueItem
