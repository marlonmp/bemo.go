package bemogo

import (
	"fmt"

	dgo "github.com/bwmarrin/discordgo"
)

type ChannelService struct {
	liknedUser string

	voiceConn *dgo.VoiceConnection
}

func NewChannelService(owner int) ChannelService {
	return ChannelService{}
}

func (cs *ChannelService) Join(session *dgo.Session, interaction *dgo.MessageCreate) {
	if cs.voiceConn != nil && cs.voiceConn.ChannelID == interaction.ChannelID {
		return
	}

	voiceState, err := session.State.VoiceState(interaction.GuildID, interaction.Author.ID)

	if err != nil {
		fmt.Println(err)
		return
	}

	cs.voiceConn, err = session.ChannelVoiceJoin(interaction.GuildID, voiceState.ChannelID, false, true)

	if err != nil {
		fmt.Println(err)
	}
}

func (cs ChannelService) Exit(session *dgo.Session, interaction *dgo.InteractionCreate) {
	if cs.voiceConn == nil {
		return
	}

	cs.voiceConn.Disconnect()
}

func (cs ChannelService) Follow(session *dgo.Session, in *dgo.VoiceStateUpdate) {
	if cs.voiceConn == nil {
		return
	}

	if cs.voiceConn.ChannelID == in.ChannelID {
		return
	}

	err := cs.voiceConn.ChangeChannel(in.ChannelID, false, true)
	if err != nil {
		fmt.Println(err)
	}
}
