package bemogo

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/dgvoice"
	dgo "github.com/bwmarrin/discordgo"
)

type interactionHandler func(session *dgo.Session, interaction *dgo.InteractionCreate)

type ChannelService struct {
	playlists map[string]*playlist
}

func NewChannelService() ChannelService {
	playlists := make(map[string]*playlist)
	return ChannelService{playlists: playlists}
}

func (cs *ChannelService) List(session *dgo.Session, interaction *dgo.InteractionCreate) {
	playlist, ok := cs.playlists[interaction.GuildID]

	if !ok {
		playlist = NewPlaylist()
		cs.playlists[interaction.GuildID] = playlist
	}

	embeds := playlist.SongsEmbed()

	if len(embeds) == 0 {
		interactionResponse(session, interaction, "There is no songs in the main playlist :/")
		return
	}

	session.InteractionRespond(interaction.Interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseChannelMessageWithSource,
		Data: &dgo.InteractionResponseData{
			Content: fmt.Sprintf("There is/are %d song(s) in the main playlist:", len(embeds)),
			Embeds:  embeds,
		},
	})
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
		interactionResponse(session, interaction, "I'm not currently in a voice channel >:P")
		return
	}

	playlist, ok := cs.playlists[interaction.GuildID]

	if !ok {
		playlist = NewPlaylist()
		cs.playlists[interaction.GuildID] = playlist
	}

	data := interaction.ApplicationCommandData()

	var songID int64

	if len(data.Options) >= 1 {
		songID = data.Options[0].IntValue() - 1
	}

	song := playlist.GetSong(songID)

	if song == nil {
		interactionResponse(session, interaction, "This song id is not valid :c")
		return
	}

	if playlist.isPlaying {
		playlist.stop <- true
		playlist.stop <- true
	}

	playlist.isPlaying = true
	playlist.currentSong = song

	session.InteractionRespond(interaction.Interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseChannelMessageWithSource,
		Data: &dgo.InteractionResponseData{
			Content: "Playing song",
			Embeds:  []*dgo.MessageEmbed{song.Embed()},
		},
	})

	dgvoice.PlayAudioFile(voiceConn, song.source, playlist.stop)

songLoop:
	for {
		song = playlist.NextSong()

		dgvoice.PlayAudioFile(voiceConn, song.source, playlist.stop)

		select {
		case <-playlist.stop:
			break songLoop
		default:
			continue
		}
	}
}

func (cs *ChannelService) TogglePause(session *dgo.Session, interaction *dgo.InteractionCreate) {
	voiceConn, ok := session.VoiceConnections[interaction.GuildID]

	if !ok {
		interactionResponse(session, interaction, "I'm not currently in a voice channel >:P")
		return
	}

	playlist, ok := cs.playlists[interaction.GuildID]

	if !ok {
		playlist = NewPlaylist()
		cs.playlists[interaction.GuildID] = playlist
	}

	song := playlist.CurrentSong()

	if song == nil {
		interactionResponse(session, interaction, "I'm not currently playing music >:P")
		return
	}

	isPlaying := playlist.TogglePause()

	if !isPlaying {
		playlist.stop <- true
		playlist.stop <- true
		interactionResponse(session, interaction, "Song paused")
		return
	}

	dgvoice.PlayAudioFile(voiceConn, song.source, playlist.stop)

	session.InteractionRespond(interaction.Interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseChannelMessageWithSource,
		Data: &dgo.InteractionResponseData{
			Content: "Playing song",
			Embeds:  []*dgo.MessageEmbed{song.Embed()},
		},
	})
}

func (cs *ChannelService) AddSong(session *dgo.Session, interaction *dgo.InteractionCreate) {
	data := interaction.ApplicationCommandData()

	videoID := data.Options[0].StringValue()

	song, err := SongFromID(videoID)

	if err != nil {
		fmt.Println("cannot create song: ", err)
		return
	}

	playlist, ok := cs.playlists[interaction.GuildID]

	if !ok {
		playlist = NewPlaylist()
		cs.playlists[interaction.GuildID] = playlist
	}

	playlist.PushSong(song)

	session.InteractionRespond(interaction.Interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseChannelMessageWithSource,
		Data: &dgo.InteractionResponseData{
			Content: "New song added",
			Embeds:  []*dgo.MessageEmbed{song.Embed()},
		},
	})
}

func (cs ChannelService) Leave(session *dgo.Session, interaction *dgo.InteractionCreate) {
	voiceConn, ok := session.VoiceConnections[interaction.GuildID]

	if !ok {
		interactionResponse(session, interaction, "Sorry, I'm not currently in a voice channel :P")
		return
	}

	playlist, ok := cs.playlists[interaction.GuildID]

	if ok && playlist.isPlaying {
		playlist.stop <- true
		playlist.stop <- true
	}

	err := voiceConn.Disconnect()

	if err != nil {
		fmt.Println("cannot disconnect from the voice connection: ", err)
		content := "It looks like there is an error trying to disconnection from the voice channel... that means I'm here forever >:D"
		interactionResponse(session, interaction, content)
		return
	}

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
		Description: "The youtube video url or the video id",
		Required:    true,
	}

	songIDOption := dgo.ApplicationCommandOption{
		Type:        dgo.ApplicationCommandOptionInteger,
		Name:        "song",
		Description: "The ID of the song in the playlist",
		Required:    false,
	}

	// timesOption := dgo.ApplicationCommandOption{
	// 	Type:        dgo.ApplicationCommandOptionInteger,
	// 	Name:        "times",
	// 	Description: "The times to do an action",
	// 	Required:    false,
	// }

	// commands
	list := dgo.ApplicationCommand{
		Name:        "list",
		Description: "List the main playlist songs",
	}

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
		Options:     []*dgo.ApplicationCommandOption{&songIDOption},
	}

	togglePause := dgo.ApplicationCommand{
		Name:        "puase",
		Description: "Toggle pause/unpause the current song playing",
	}

	addSong := dgo.ApplicationCommand{
		Name:        "add",
		Description: "Add a song in the main playlist",
		Options:     []*dgo.ApplicationCommandOption{&songYTOption},
	}

	// popSong := dgo.ApplicationCommand{
	// 	Name:        "pop",
	// 	Description: "Remove the last song added or the songID passed",
	// 	Options:     []*dgo.ApplicationCommandOption{&songIDOption},
	// }

	// jumpPrevious := dgo.ApplicationCommand{
	// 	Name:        "previous",
	// 	Description: "Play the previous song according to the times specified in the params (by default 1)",
	// 	Options:     []*dgo.ApplicationCommandOption{&timesOption},
	// }

	// jumpNext := dgo.ApplicationCommand{
	// 	Name:        "next",
	// 	Description: "Play the next song according to the times specified in the params (by default 1)",
	// 	Options:     []*dgo.ApplicationCommandOption{&timesOption},
	// }

	// bucleSong := dgo.ApplicationCommand{
	// 	Name:        "buble",
	// 	Description: "Repeat the current playing song, one or more times (by default 1)",
	// 	Options:     []*dgo.ApplicationCommandOption{&timesOption},
	// }

	// restartPlaylist := dgo.ApplicationCommand{
	// 	Name:        "restart",
	// 	Description: "Restart the main playlist status",
	// }

	// shufflePlaylist := dgo.ApplicationCommand{
	// 	Name:        "shuffle",
	// 	Description: "Shuffle the main playlist",
	// }

	commands := []dgo.ApplicationCommand{list, join, leave, playSong, togglePause, addSong}

	handlers := []interactionHandler{cs.List, cs.Join, cs.Leave, cs.PlaySong, cs.TogglePause, cs.AddSong}

	return commands, handlers
}

func (cs ChannelService) CommandHandler(session *dgo.Session, interaction *dgo.InteractionCreate) {
	if interaction.ApplicationCommandData().Type() != dgo.InteractionApplicationCommand {
		return
	}

	data := interaction.ApplicationCommandData()

	commands, handlers := cs.GetCommandsHandlers()

	for i, command := range commands {
		if data.Name == command.Name {
			println("enter in command: ", command.Name)
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
