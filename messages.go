package bemogo

import (
	dgo "github.com/bwmarrin/discordgo"
)

func interactionResponse(session *dgo.Session, interaction *dgo.InteractionCreate, content string) error {
	return session.InteractionRespond(interaction.Interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseChannelMessageWithSource,
		Data: &dgo.InteractionResponseData{Content: content},
	})
}
