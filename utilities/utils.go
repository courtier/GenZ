package utilities

import (
	"os"
	"path/filepath"
)

//help messages
const (
	HelpMessage string = "```GenZ Bot - Brought To You By Goob, Potato And Courtier```\n" +
		"*Commands (need to start the message with {prefix}):*\n\n" +
		"***help*** -> shows you what the bot can do\n\n" +
		"***broken*** (use in DMs) -> reports the account as broken and it gets removed from the database\n" +
		"_note: you will get banned from generating accounts for some time if you false report_\n\n" +
		"***works*** (use in DMs) -> reports the account as working\n" +
		"_please dont abuse broken/works commands_"

	AdminHelpMessage string = "```GenZ Bot - GenZ on top!```\n" +
		"*Commands (need to start with {prefix}):*\n\n" +
		"***help*** -> shows you what the bot can do\n\n" +
		"***adminhelp*** -> shows you what admins can do\n\n" +
		"***generate*** -> sends account generation message to the channel and people can react to get accounts\n\n" +
		"***add <category>*** -> saves the next message you send to the database\n\n" +
		"***remove <account id>*** -> removes account id\n\n" +
		"***remove <category name>*** -> removes an entire category"
)

//Contains contains and dindex of string in string array
func Contains(array []string, element string) (contains bool, index int) {
	for i, item := range array {
		if item == element {
			return true, i
		}
	}
	return false, 0
}

//GetCurrentPath get path GenZ is running out of
func GetCurrentPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "./"
	}
	return dir
}
