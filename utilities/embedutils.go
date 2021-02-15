package utilities

import (
	"strings"

	embedManager "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/courtier/eggshell"
)

const (
	pinkColor      int = 16722332
	defTitle           = "GenZ"
	defDescription     = "React with one of the letters to receive an account"
	defFooter          = "Don't change the password of the accounts"
)

//Alphabet alphabet as string array
var Alphabet = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}

//AlphabetEmojis emojis for da alphabet
var AlphabetEmojis = []string{"🇦", "🇧", "🇨", "🇩", "🇪", "🇫", "🇬", "🇭", "🇮", "🇯", "🇰", "🇱", "🇲", "🇳", "🇴", "🇵", "🇶", "🇷", "🇸", "🇹", "🇺", "🇻", "🇼", "🇽", "🇾", "🇿"}

//GetEmbedID return embed id
func GetEmbedID() (messageID string) {
	return
}

//RefreshEmbed refresh embed
func RefreshEmbed(db *eggshell.Driver) (embed *discordgo.MessageEmbed, numberOfCategories int) {
	//add field: Categories:````
	categoryList := GetAllCategories(db)
	if len(categoryList) < 1 {
		return
	}
	fieldValue := "```\n"
	var size int = 0
	for index, category := range categoryList {
		fieldValue += strings.ToUpper(Alphabet[index]) + " -> " + strings.Title(category) + "\n"
		size++
	}
	fieldValue += "```"
	filledEmbed := embedManager.NewEmbed().SetTitle(defTitle).SetDescription(defDescription).SetColor(pinkColor).AddField("Categories", fieldValue)
	//filledEmbed := defaultEmbed.AddField("Categories", fieldValue)
	return filledEmbed.MessageEmbed, size
}

//ParseCategoryFromEmbedd parse category from embed using emoji
func ParseCategoryFromEmbedd(emoji string, embed *discordgo.MessageEmbed) (category string) {
	field := embed.Fields[0]
	lines := strings.Split(field.Value, "\n")
	categoryFound := ""
	_, emojiIndex := Contains(AlphabetEmojis, emoji)
	categoryPrefix := strings.ToUpper(Alphabet[emojiIndex]) + " -> "
	for _, line := range lines {
		contains := strings.Contains(line, categoryPrefix)
		if contains {
			categoryFound = strings.Split(line, "-> ")[1]
		}
	}
	return categoryFound
}

//CreateAccountEmbed create embed from acccount
func CreateAccountEmbed(category string, account Account) (createdEmbed *discordgo.MessageEmbed) {
	var accountEmbed *embedManager.Embed = embedManager.NewEmbed().
		SetTitle(strings.Title(category) + " Account").
		SetDescription("React with the red X if the account is not working!").
		SetFooter("Don't change the password of the accounts\nIf you're on mobile react with the phone emoji").
		SetColor(pinkColor)
	if len(strings.Split(account.Info, ":")) > 1 {
		password := strings.Split(account.Info, ":")[1]
		password = strings.Split(password, " ")[0]
		accountEmbed.
			AddField("Email/Username", "```"+strings.Split(account.Info, ":")[0]+"```").
			AddField("Password", "```"+password+"```").
			AddField("ID", "```"+account.ID+"```")
	} else {
		accountEmbed.
			AddField("Information", "```"+account.Info+"```").
			AddField("ID", "```"+account.ID+"```")
	}
	return accountEmbed.MessageEmbed
}

//CreateAccountMessage create message from acccount
func CreateAccountMessage(category string, account Account) (createdMessage string) {
	var accountMessage string = ""

	if len(strings.Split(account.Info, ":")) > 1 {
		password := strings.Split(account.Info, ":")[1]
		password = strings.Split(password, " ")[0]
		accountMessage = "***Email/Username***\n" + "```" + strings.Split(account.Info, ":")[0] + "```\n" +
			"***Password***" + "```" + password + "```\n" +
			"***ID***" + "```" + account.ID + "```"
	} else {
		accountMessage = "Information\n" + account.Info + "\n" +
			"ID\n" + account.ID
	}
	return accountMessage
}

//CreateAccountMessageSeparate create messages from acccount
func CreateAccountMessageSeparate(category string, account Account) (createdMessage []string) {
	accountMessages := []string{}

	if len(strings.Split(account.Info, ":")) > 1 {

		accountMessages = append(accountMessages, "***Email/Username***")
		accountMessages = append(accountMessages, strings.Split(account.Info, ":")[0])

		password := strings.Split(account.Info, ":")[1]
		password = strings.Split(password, " ")[0]

		accountMessages = append(accountMessages, "***Password***")
		accountMessages = append(accountMessages, password)

		accountMessages = append(accountMessages, "***ID***")
		accountMessages = append(accountMessages, account.ID)

	} else {
		accountMessages = append(accountMessages, "***Account Information***")
		accountMessages = append(accountMessages, account.Info)

		accountMessages = append(accountMessages, "***ID***")
		accountMessages = append(accountMessages, account.ID)
	}

	return accountMessages
}

//GenerateNewEmbed refreshes embed
func GenerateNewEmbed(s *discordgo.Session, db *eggshell.Driver, channelID string) {
	foundEmbed := RetrieveEmbed(db)
	s.ChannelMessageDelete(foundEmbed.ChannelID, foundEmbed.MessageID)
	DeleteEmbed(db)
	newEmbed, numberOfCategories := RefreshEmbed(db)
	sendChannelID := channelID
	if newEmbed == nil || numberOfCategories == 0 {
		s.ChannelMessageSend(channelID, "couldnt find any categories")
		return
	}
	if foundEmbed.ChannelID != "" {
		sendChannelID = foundEmbed.ChannelID
	}
	msg, err := s.ChannelMessageSendEmbed(sendChannelID, newEmbed)
	if err != nil {
		s.ChannelMessageSend(channelID, "error sending embed")
	}

	for i := 0; i < numberOfCategories; i++ {
		s.MessageReactionAdd(msg.ChannelID, msg.ID, AlphabetEmojis[i])
	}
	if !SaveEmbed(db, msg.ID, msg.ChannelID) {
		s.ChannelMessageSend(msg.ChannelID, "error saving embed")
	}

}
