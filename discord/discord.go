package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/masgustavos/alertmanager-discord/alertmanager"
	"github.com/masgustavos/alertmanager-discord/config"
)

// SendAlerts deals with the macro logic of sending alerts to Discord Channels
func SendAlerts(
	discordChannelName string,
	alertmanagerBody alertmanager.MessageBody,
	configs config.Config) error {

	alertmanagerBodyInfo := alertmanager.ExtractBodyInfo(alertmanagerBody, configs)

	discordChannel, err := getDiscordChannel(discordChannelName, configs)
	if err != nil {
		return fmt.Errorf("discord.SendAlerts: Error trying to get Discord Channel \n%+v", err)
	}

	if alertmanager.CheckIfHasOnlySeveritiesToIgnoreWhenAlone(
		alertmanagerBodyInfo.CountBySeverity,
		discordChannel, configs) {

		return fmt.Errorf(
			`discord.SendAlerts: There are only alerts with severities to be ignored, message not sent.
			Severity Count: %+v`,
			alertmanagerBodyInfo.CountBySeverity)

	}

	discordMessage, err := createDiscordMessage(alertmanagerBodyInfo, discordChannel, configs)
	if err != nil {
		return fmt.Errorf("discord.SendAlerts: Error trying to create Discord Message \n%+v", err)
	}

	jsonDiscordMessage, err := json.Marshal(discordMessage)
	if err != nil {
		return fmt.Errorf("discord.SendAlerts: Error Marshaling Discord Message \n%+v", err)
	}

	r, err := http.Post(
		discordChannel.WebhookURL,
		"application/json",
		bytes.NewReader(jsonDiscordMessage))

	if err != nil {
		return fmt.Errorf("discord.SendAlerts: Error Posting alert to Discord \n%+v", err)
	}

	defer r.Body.Close()

	if r.StatusCode != 204 {
		return fmt.Errorf(
			`discord.SendAlerts: Problem with Post, status code is not 204.
			StatusCode: %d, Message: %s`,
			r.StatusCode, r.Status)
	}

	return nil
}

func getDiscordChannel(
	discordChannelName string,
	configs config.Config) (config.DiscordChannel, error) {

	discordChannel, ok := configs.DiscordChannels[discordChannelName]
	if ok {
		return discordChannel, nil
	}

	err := fmt.Errorf(
		"discord.getDiscordChannel: The discordChannel %s could not be found",
		discordChannelName)
	return config.DiscordChannel{}, err
}

func createDiscordMessage(
	alertmanagerBodyInfo alertmanager.MessageBodyInfo,
	discordChannel config.DiscordChannel,
	configs config.Config) (message WebhookParams, err error) {

	content := fmt.Sprintf(
		"Firing: %d  |  Resolved: %d",
		alertmanagerBodyInfo.FiringCount, alertmanagerBodyInfo.ResolvedCount)

	handleMentions(alertmanagerBodyInfo, &content, discordChannel, configs)

	firingEmbeds, err := createDiscordMessageEmbeds(alertmanagerBodyInfo.FiringAlertsGroupedByName,
		"firing", configs)
	if err != nil {
		err = fmt.Errorf("discord.createDiscordMessage: Error creating firingEmbeds\n%+v", err)
		return WebhookParams{}, err
	}

	resolvedEmbeds, err := createDiscordMessageEmbeds(alertmanagerBodyInfo.ResolvedAlertsGroupedByName,
		"resolved", configs)
	if err != nil {
		err = fmt.Errorf("discord.createDiscordMessage: Error creating resolvedEmbeds %+v", err)
		return WebhookParams{}, err
	}

	embeds := append(firingEmbeds, resolvedEmbeds...)

	return WebhookParams{
		Content:   content,
		Embeds:    embeds,
		Username:  configs.Username,
		AvatarURL: configs.AvatarURL}, nil
}

func handleMentions(
	alertmanagerBodyInfo alertmanager.MessageBodyInfo,
	content *string,
	discordChannel config.DiscordChannel,
	configs config.Config) {

	var severitiesToMention []string

	// Channels can override global severitiesToMention
	if len(discordChannel.SeveritiesToMention) > 0 {
		severitiesToMention = discordChannel.SeveritiesToMention
	} else if len(configs.SeveritiesToMention) > 0 {
		severitiesToMention = configs.SeveritiesToMention
	}

	shouldMentionBySeverity := checkIfShouldMentionBySeverity(severitiesToMention, alertmanagerBodyInfo, configs)
	shouldMentionByFiringCount := checkIfShouldMentionByFiringCount(alertmanagerBodyInfo, configs)

	if shouldMentionBySeverity || shouldMentionByFiringCount {
		addRolesToEmbedContent(content, discordChannel, configs)
	}

}

func checkIfShouldMentionBySeverity(
	severitiesToMention []string,
	alertmanagerBodyInfo alertmanager.MessageBodyInfo,
	configs config.Config) bool {

	for _, severityToMention := range severitiesToMention {
		if alertmanagerBodyInfo.CountBySeverity[severityToMention] > 0 {
			return true
		}
	}

	return false
}

func checkIfShouldMentionByFiringCount(
	alertmanagerBodyInfo alertmanager.MessageBodyInfo,
	configs config.Config) bool {

	if configs.FiringCountToMention > 0 {
		if alertmanagerBodyInfo.FiringCount >= configs.FiringCountToMention {
			return true
		}
	}

	return false

}

func addRolesToEmbedContent(
	content *string,
	discordChannel config.DiscordChannel,
	configs config.Config) {

	// Channels can override rolesToMention
	if len(discordChannel.RolesToMention) > 0 {
		*content = *content + "    " + strings.Join(discordChannel.RolesToMention, " ")
	} else {
		*content = *content + "    " + strings.Join(configs.RolesToMention, " ")
	}

}

func createDiscordMessageEmbeds(
	alertsGroupedByName alertmanager.AlertsGroupedByLabel,
	status string,
	configs config.Config) ([]MessageEmbed, error) {

	embedQueue := []EmbedQueueItem{}

	for _, alerts := range alertsGroupedByName {

		embed := MessageEmbed{}

		if alerts[0].Annotations.Summary != "" {
			embed.Title = fmt.Sprintf("%s\n", alerts[0].Annotations.Summary)
		} else {
			embed.Title = fmt.Sprintf("%s\n", alerts[0].Labels["alertname"])
		}

		embed.URL = alerts[0].GeneratorURL

		for _, alert := range alerts {
			if alert.Annotations.Description == "" {
				embed.Description = "```No description provided```"
				break
			}
			embed.Description = embed.Description + fmt.Sprintf("```%s```\n", alert.Annotations.Description)
		}

		priority, err := handleEmbedAppearance(&embed, status, alerts[0], configs)
		if err != nil {
			err = fmt.Errorf(
				`discord.createDiscordMessageEmbeds:
				Couldn't handle embed appearance for embed %+v and alert %+v: \n%+v`,
				embed, alerts[0], err)
			return []MessageEmbed{}, err
		}

		embedQueueItem := EmbedQueueItem{
			Embed:    embed,
			Priority: priority,
		}

		embedQueue = append(embedQueue, embedQueueItem)
	}

	sort.Slice(embedQueue[:], func(i, j int) bool {
		return embedQueue[i].Priority > embedQueue[j].Priority
	})

	embedsOrderedByPriority := []MessageEmbed{}

	for _, embedQueueItem := range embedQueue {
		embedsOrderedByPriority = append(embedsOrderedByPriority, embedQueueItem.Embed)
	}

	return embedsOrderedByPriority, nil
}

func handleEmbedAppearance(
	embed *MessageEmbed, status string,
	alert alertmanager.Alert,
	configs config.Config) (priority int, err error) {

	if status == "resolved" {
		embed.Color = configs.Status["resolved"].Color
		embed.Title = fmt.Sprintf("\n%s %s", configs.Status["resolved"].Emoji, embed.Title)
		return 0, nil
	} else if status == "firing" {
		switch configs.MessageType {
		case "status":
			embed.Color = configs.Status["firing"].Color
			embed.Title = fmt.Sprintf("\n%s %s", configs.Status["firing"].Emoji, embed.Title)
			return 0, nil
		case "severity":
			severityAppearance := handleEmbedSeverity(embed, alert, configs)
			return severityAppearance.Priority, nil
		default:
			return 0, fmt.Errorf(
				"discord.handleEmbedAppearance: No matching message type for %s",
				configs.MessageType)
		}
	}

	return 0, nil
}

func handleEmbedSeverity(embed *MessageEmbed, alert alertmanager.Alert, configs config.Config) config.SeverityAppearance {
	severity, ok := alert.Labels[configs.Severity.Label]
	var SeverityAppearance config.SeverityAppearance
	if ok {
		SeverityAppearance, ok = configs.Severity.Values[severity]
		if !ok {
			SeverityAppearance = configs.Severity.Values["unknown"]
		}
		embed.Title = fmt.Sprintf("\n%s %s", SeverityAppearance.Emoji, embed.Title)
		embed.Color = SeverityAppearance.Color
	}
	return SeverityAppearance
}
