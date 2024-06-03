package bemogo

import (
	"fmt"

	dgo "github.com/bwmarrin/discordgo"
)

type interactionHandler func(session *dgo.Session, interaction *dgo.InteractionCreate)

type ChannelService struct {
	linkedUser map[string]string
}

func NewChannelService(owner int) ChannelService {
	return ChannelService{}
}

func (cs *ChannelService) Join(session *dgo.Session, interaction *dgo.InteractionCreate) {
	voiceState, err := session.State.VoiceState(interaction.GuildID, interaction.Member.User.ID)

	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = session.ChannelVoiceJoin(interaction.GuildID, voiceState.ChannelID, false, true)

	if err != nil {
		fmt.Println(err)
	}

	session.InteractionRespond(interaction.Interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseChannelMessageWithSource,
		Data: &dgo.InteractionResponseData{
			Content: fmt.Sprintf("Joining in <#%s>", voiceState.ChannelID),
		},
	})
}

func (cs ChannelService) Leave(session *dgo.Session, interaction *dgo.InteractionCreate) {
	if cs.voiceConn == nil {
		return
	}

	voiceConn, ok := session.VoiceConnections[interaction.GuildID]

	if !ok {
		return
	}

	voiceConn.Disconnect()

	session.InteractionRespond(interaction.Interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseChannelMessageWithSource,
		Data: &dgo.InteractionResponseData{
			Content: "Bye bye!!",
		},
	})
}

func (cs ChannelService) Follow(session *dgo.Session, voiceState *dgo.VoiceStateUpdate) {
	println("voice status chanaged")

	var err error

	if cs.voiceConn != nil && cs.voiceConn.ChannelID == voiceState.ChannelID {
		return
	}

	_, err = session.ChannelVoiceJoin(voiceState.GuildID, voiceState.ChannelID, false, true)

	if err != nil {
		fmt.Println(err)
	}
}

func (cs ChannelService) GetCommandsHandlers() ([]dgo.ApplicationCommand, []interactionHandler) {
	// options
	user := dgo.ApplicationCommandOption{
		Type:        dgo.ApplicationCommandOptionUser,
		Name:        "user",
		Description: "The user who will be linked to",
		Required:    false,
	}

	songYTOption := dgo.ApplicationCommandOption{
		Type:        dgo.ApplicationCommandOptionString,
		Name:        "song-url",
		Description: "The youtube video url",
	}

	songIDOption := dgo.ApplicationCommandOption{
		Type:        dgo.ApplicationCommandOptionInteger,
		Name:        "song",
		Description: "The ID of the song in the playlist",
		Required:    false,
	}

	timesOption := dgo.ApplicationCommandOption{
		Type:        dgo.ApplicationCommandOptionInteger,
		Name:        "times",
		Description: "The times to do an action",
		Required:    false,
	}

	// commands
	join := dgo.ApplicationCommand{
		Name:        "join",
		Description: "Join in the user voice channel",
	}

	leave := dgo.ApplicationCommand{
		Name:        "leave",
		Description: "Leave from the current voice channel",
	}

	linkUser := dgo.ApplicationCommand{
		Name:        "link",
		Description: "Link bemogo to an user, bemo will follow the user throw the guild voices channels",
		Options:     []*dgo.ApplicationCommandOption{&user},
	}

	addSong := dgo.ApplicationCommand{
		Name:        "add",
		Description: "Add a song in the main playlist",
		Options:     []*dgo.ApplicationCommandOption{&songYTOption},
	}

	popSong := dgo.ApplicationCommand{
		Name:        "pop",
		Description: "Remove the last song added or the songID passed",
		Options:     []*dgo.ApplicationCommandOption{&songIDOption},
	}

	togglePause := dgo.ApplicationCommand{
		Name:        "puase",
		Description: "Toggle pause/unpause the current song playing",
	}

	jumpPrevious := dgo.ApplicationCommand{
		Name:        "previous",
		Description: "Play the previous song according to the times specified in the params (by default 1)",
		Options:     []*dgo.ApplicationCommandOption{&timesOption},
	}

	jumpNext := dgo.ApplicationCommand{
		Name:        "next",
		Description: "Play the next song according to the times specified in the params (by default 1)",
		Options:     []*dgo.ApplicationCommandOption{&timesOption},
	}

	bucleSong := dgo.ApplicationCommand{
		Name:        "buble",
		Description: "Repeat the current playing song, one or more times (by default 1)",
		Options:     []*dgo.ApplicationCommandOption{&timesOption},
	}

	restartPlaylist := dgo.ApplicationCommand{
		Name:        "restart",
		Description: "Restart the main playlist status",
	}

	shufflePlaylist := dgo.ApplicationCommand{
		Name:        "shuffle",
		Description: "Shuffle the main playlist",
	}

	commands := []dgo.ApplicationCommand{join, leave, linkUser, addSong, popSong, togglePause, jumpPrevious, jumpNext, bucleSong, restartPlaylist, shufflePlaylist}

	handlers := []interactionHandler{cs.Join, cs.Leave}

	return commands, handlers
}

func (cs ChannelService) CommandHandler(session *dgo.Session, interaction *dgo.InteractionCreate) {
	if interaction.ApplicationCommandData().Type() != dgo.InteractionApplicationCommand {
		return
	}

	commands, handlers := cs.GetCommandsHandlers()

	for i, command := range commands {
		if interaction.ApplicationCommandData().Name == command.Name {
			handlers[i](session, interaction)
			return
		}
	}
}

func (cs ChannelService) AddCommands(session *dgo.Session, ready *dgo.Ready) {
	guildsID := getGuildsWhitelist()

	commands, _ := cs.GetCommandsHandlers()

	for _, guildID := range guildsID {
		for _, command := range commands {
			_, err := session.ApplicationCommandCreate(session.State.User.ID, guildID, &command)

			if err != nil {
				fmt.Println(`cannot register the command "`, command.Name, `": `, err)
				continue
			}
		}
	}

	session.AddHandler(cs.CommandHandler)
}

func (cs ChannelService) RemoveCommands(session *dgo.Session) {
	guildsID := getGuildsWhitelist()

	for _, guildID := range guildsID {
		commands, err := session.ApplicationCommands(session.State.User.ID, guildID)

		if err != nil {
			fmt.Println("cannot get the application commands: ", err)
			continue
		}

		for _, command := range commands {
			err := session.ApplicationCommandDelete(session.State.User.ID, guildID, command.ID)

			if err != nil {
				fmt.Println(`cannot delete command "`, command.Name, `": `, err)
				continue
			}
		}
	}
}
