package main

import (
	"bufio"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/courtier/eggshell"
	"github.com/gookit/color"
	"github.com/pelletier/go-toml"

	"genz/utilities"
)

var (
	config    *toml.Tree
	isRunning bool = true
	db        *eggshell.Driver
)

func main() {

	db = utilities.LoadDB()

	color.Blue.Println("starting bot")

	config = loadConfig()

	if len(config.Get("prefix").(string)) == 0 {
		config.Set("prefix", ",")
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + config.Get("token").(string))
	if err != nil {
		color.Red.Println("error starting bot, ", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)
	dg.AddHandler(reactionAdd)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsDirectMessages | discordgo.IntentsDirectMessageReactions)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		color.Red.Println("error opening connection, ", err)
		return
	}

	dg.UpdateStatus(0, config.Get("playing").(string))

	// Wait here until CTRL-C or other term signal is received.
	color.Green.Println("bot has started, ctrl-c to exit")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

//called errtime message received
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// ignore all bot messages cause we dont care about em
	if m.Author.Bot {
		return
	}

	//check if message starts with prefix
	if strings.HasPrefix(m.Content, config.Get("prefix").(string)) {

		if !isRunning && m.Content != config.Get("prefix").(string)+"enable" {
			s.ChannelMessageSend(m.ChannelID, "bot is disabled!")
			return
		}

		cleanCommand := strings.Replace(m.Content, config.Get("prefix").(string), "", 1)

		//not-admin available commands
		if cleanCommand == "help" {

			formatted := strings.Replace(utilities.HelpMessage, "{prefix}", config.Get("prefix").(string), 1)
			s.ChannelMessageSend(m.ChannelID, formatted)

		}

		//admin comments

		if strings.Contains(config.String(), m.Author.ID) {

			if cleanCommand == "adminhelp" {

				formatted := strings.Replace(utilities.AdminHelpMessage, "{prefix}", config.Get("prefix").(string), 1)
				s.ChannelMessageSend(m.ChannelID, formatted)

			} else if cleanCommand == "disable" || cleanCommand == "enable" {

				//if disabled, bot will not respond to commands
				isRunning = cleanCommand == "enable"
				s.ChannelMessageSend(m.ChannelID, "bot is enabled: "+strconv.FormatBool(isRunning))

			} else if cleanCommand == "generate" {

				utilities.GenerateNewEmbed(s, db, m.ChannelID)

			} else if cleanCommand == "removeembed" {

				foundEmbed := utilities.RetrieveEmbed(db)
				s.ChannelMessageDelete(foundEmbed.ChannelID, foundEmbed.MessageID)
				utilities.DeleteEmbed(db)

			} else if strings.HasPrefix(cleanCommand, "add") {

				category := strings.Split(cleanCommand, " ")[1]
				category = strings.Split(category, "\n")[0]

				if len(strings.Split(cleanCommand, " ")) == 3 && strings.HasPrefix(strings.Split(cleanCommand, " ")[2], "http") {
					accountLines := loadAccountsFromLink(strings.Split(cleanCommand, " ")[2])
					if len(accountLines) < 1 {
						s.ChannelMessageSend(m.ChannelID, "couldnt find any accounts in the message")
						return
					}
					if utilities.InsertAccounts(db, category, accountLines) {
						s.ChannelMessageSend(m.ChannelID, "inserted "+strconv.Itoa(len(accountLines))+" accounts into the database")
					}

					utilities.GenerateNewEmbed(s, db, m.ChannelID)
					return
				}

				accountLines := strings.Split(cleanCommand, "\n")
				linesLength := len(accountLines) - 1
				if linesLength < 1 {
					s.ChannelMessageSend(m.ChannelID, "couldnt find any accounts in the message")
					return
				}
				accountLines = accountLines[1 : linesLength+1]

				if utilities.InsertAccounts(db, category, accountLines) {
					s.ChannelMessageSend(m.ChannelID, "inserted "+strconv.Itoa(linesLength)+" accounts into the database")
				}

				utilities.GenerateNewEmbed(s, db, m.ChannelID)

			} else if strings.HasPrefix(cleanCommand, "remove") {

				category := strings.Split(cleanCommand, " ")[1]
				worked := utilities.RemoveCategory(db, category)
				if worked {
					s.ChannelMessageSend(m.ChannelID, "removed "+category)
					utilities.GenerateNewEmbed(s, db, m.ChannelID)
				} else {
					s.ChannelMessageSend(m.ChannelID, "couldnt remove "+category)
				}

			}

		}

	}

}

func reactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {

	user, _ := s.User(m.UserID)
	if user.Bot {
		return
	}

	if !isRunning {
		return
	}

	self, _ := s.User("@me")
	message, err := s.ChannelMessage(m.ChannelID, m.MessageID)
	if message == nil || err != nil {
		return
	}
	if message.Author.ID == self.ID {

		if len(message.Embeds) == 1 {
			foundEmbed := message.Embeds[0]
			if foundEmbed.Title == "GenZ" {

				category := utilities.ParseCategoryFromEmbedd(m.Emoji.Name, foundEmbed)
				if category != "" {

					gennedAccount, err := utilities.RetrieveGennedAccount(db, m.UserID)

					timeNow := time.Now().Unix()
					status := 0

					if err != nil || gennedAccount.Time == 0 {
						status = 1
					} else if timeNow-gennedAccount.Time >= config.Get("cooldown-duration").(int64) {
						status = 2
					} else if timeNow-gennedAccount.Time < config.Get("cooldown-duration").(int64) {
						privChannel, _ := s.UserChannelCreate(m.UserID)
						s.ChannelMessageSend(privChannel.ID, strconv.Itoa(int(30-(timeNow-gennedAccount.Time)))+" seconds left on your cooldown")
						return
					}

					if status > 0 {
						account := utilities.ReceiveAccount(db, category)
						privChannel, err := s.UserChannelCreate(m.UserID)
						if err != nil {
							color.Red.Println("error while retrieving private channel", err)
							return
						}
						msg, err := s.ChannelMessageSendEmbed(privChannel.ID, utilities.CreateAccountEmbed(category, account))
						if err != nil {
							color.Red.Println("error while sending message to private channel", err)
							return
						}
						s.MessageReactionAdd(msg.ChannelID, msg.ID, "📞")
						s.MessageReactionAdd(msg.ChannelID, msg.ID, "❌")
						gennedAccount = utilities.GennedAccount{UserID: m.UserID, AccountID: account.ID, Time: timeNow}
						utilities.DeleteGennedAccount(db, m.UserID)
						defer utilities.SaveGennedAccount(db, gennedAccount)
					}

				}

			} else if strings.HasSuffix(foundEmbed.Title, " Account") {

				if m.Emoji.Name == "📞" {

					category := strings.ToLower(strings.ReplaceAll(foundEmbed.Title, " Account", ""))
					uuid := strings.ReplaceAll(foundEmbed.Fields[len(foundEmbed.Fields)-1].Value, "`", "")

					account := utilities.GetAccountFromUUID(db, category, uuid)

					messages := utilities.CreateAccountMessageSeparate(category, account)

					for _, msg := range messages {
						s.ChannelMessageSend(m.ChannelID, msg)
					}

				} else if m.Emoji.Name == "❌" {

					category := strings.ToLower(strings.ReplaceAll(foundEmbed.Title, " Account", ""))
					uuid := strings.ReplaceAll(foundEmbed.Fields[len(foundEmbed.Fields)-1].Value, "`", "")

					maxReports := int(config.Get("remove-after-not-working").(int64))
					if worked := utilities.ReportAccount(db, category, uuid, m.UserID, maxReports); worked {
						s.ChannelMessageSend(m.ChannelID, "reported account successfully")
					} else {
						s.ChannelMessageSend(m.ChannelID, "couldnt report account")
					}

				}

			}

		}

	}

}

func loadAccountsFromLink(link string) []string {
	resp, err := http.Get(link)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	var accountLines []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		accountLines = append(accountLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil
	}

	return accountLines
}

func loadConfig() (config *toml.Tree) {
	dir := utilities.GetCurrentPath()
	if _, err := os.Stat(dir + "/configuration.toml"); os.IsNotExist(err) {
		color.Red.Println("couldnt find configuration.toml, quitting")
		os.Exit(1)
	} else {
		config, err := toml.LoadFile(dir + "/configuration.toml")
		if err != nil {
			color.Red.Println("error while loading configuration.toml, quitting")
			os.Exit(2)
		}
		return config
	}
	return
}
