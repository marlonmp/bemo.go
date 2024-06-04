package bemogo

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/dgvoice"
	dgo "github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"
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

	if errors.Is(err, dgo.ErrStateNotFound) {
		interactionResponse(session, interaction, "Ok genius, you're not in a voice channel, how do you want me to join in your voice channel? 7-7")
		return
	}

	if err != nil {
		fmt.Println("error getting voice state: ", err)

		interactionResponse(session, interaction, "It looks like there is an error trying to get the user voice state :/")
		return
	}

	voiceConn, err := session.ChannelVoiceJoin(interaction.GuildID, voiceState.ChannelID, false, true)

	if err != nil {
		fmt.Println("error joining channel voice: ", err)

		interactionResponse(session, interaction, "It looks like there is an error when joining in the voice channel :/")
		return
	}

	content := fmt.Sprintf("I'm joining in <#%s> :3", voiceConn.ChannelID)

	interactionResponse(session, interaction, content)
}

func (cs *ChannelService) PlaySong(session *dgo.Session, interaction *dgo.InteractionCreate) {
	voiceConn, ok := session.VoiceConnections[interaction.GuildID]

	if !ok {
		return
	}

	data := interaction.ApplicationCommandData()

	videoID := data.Options[0]

	client := youtube.Client{}

	video, err := client.GetVideo(videoID.StringValue())

	if err != nil {
		fmt.Println("cannot get video info: ", err)
		return
	}

	formats := video.Formats.WithAudioChannels()

	url, err := client.GetStreamURL(video, &formats[0])

	// stream, _, err := client.GetStream(video, &formats[0])

	if err != nil {
		fmt.Println("cannot get stream: ", err)
		return
	}

	// defer stream.Close()

	dgvoice.PlayAudioFile(voiceConn, url, make(<-chan bool))

	content := fmt.Sprintf("Joining in <#%s>", voiceConn.ChannelID)

	interactionResponse(session, interaction, content)
}

func (cs ChannelService) Leave(session *dgo.Session, interaction *dgo.InteractionCreate) {
	voiceConn, ok := session.VoiceConnections[interaction.GuildID]

	if !ok {
		interactionResponse(session, interaction, "I'm not currently in a voice channel :P")
		return
	}

	voiceConn.Disconnect()

	interactionResponse(session, interaction, "Ok, I'm leaving >:c")
}

func (cs ChannelService) Follow(session *dgo.Session, voiceState *dgo.VoiceStateUpdate) {
	if session.State.User.ID == voiceState.UserID {
		return
	}

	voiceConn, ok := session.VoiceConnections[voiceState.GuildID]

	if !ok {
		return
	}

	if voiceConn != nil && voiceConn.ChannelID == voiceState.ChannelID {
		return
	}

	_, err := session.ChannelVoiceJoin(voiceState.GuildID, voiceState.ChannelID, false, true)

	if err != nil {
		fmt.Println(err)
	}
}

func (cs ChannelService) GetCommandsHandlers() ([]dgo.ApplicationCommand, []interactionHandler) {
	// options
	// user := dgo.ApplicationCommandOption{
	// 	Type:        dgo.ApplicationCommandOptionUser,
	// 	Name:        "user",
	// 	Description: "The user who will be linked to",
	// 	Required:    false,
	// }

	songYTOption := dgo.ApplicationCommandOption{
		Type:        dgo.ApplicationCommandOptionString,
		Name:        "song-url",
		Description: "The youtube video url",
		Required:    true,
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

	// linkUser := dgo.ApplicationCommand{
	// 	Name:        "link",
	// 	Description: "Link bemogo to an user, bemo will follow the user throw the guild voices channels",
	// 	Options:     []*dgo.ApplicationCommandOption{&user},
	// }

	playSong := dgo.ApplicationCommand{
		Name:        "play",
		Description: "Play a song in the main playlist",
		Options:     []*dgo.ApplicationCommandOption{&songYTOption},
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

	commands := []dgo.ApplicationCommand{join, leave, playSong, addSong, popSong, togglePause, jumpPrevious, jumpNext, bucleSong, restartPlaylist, shufflePlaylist}

	handlers := []interactionHandler{cs.Join, cs.Leave, cs.PlaySong}

	return commands, handlers
}

func (cs ChannelService) CommandHandler(session *dgo.Session, interaction *dgo.InteractionCreate) {
	if interaction.ApplicationCommandData().Type() != dgo.InteractionApplicationCommand {
		return
	}

	data := interaction.ApplicationCommandData()

	commands, handlers := cs.GetCommandsHandlers()

	for i, command := range commands {
		println(data.Name, command.Name)

		if data.Name == command.Name {
			handlers[i](session, interaction)
			return
		}
	}
}

func (cs ChannelService) AddCommands(session *dgo.Session, ready *dgo.Ready) {
	println("log: adding commans into the guilds...")

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

	println("log: all commands added, Bemo is ready to play music")
}

func (cs ChannelService) RemoveCommands(session *dgo.Session) {
	println("log: removing commands...")

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

	println("log: commands removed successfully")
}
