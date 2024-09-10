package bemogo

import (
	"os"
	"strings"
)

func getGuildsWhitelist() []string {
	whitelist := os.Getenv("DISCORD_GUILD_ID_WHITELIST")

	return strings.Split(whitelist, ",")
}

func getChannelsWhitelist() []string {
	whitelist := os.Getenv("DISCORD_CHANNEL_ID_WHITELIST")

	return strings.Split(whitelist, ",")
}
