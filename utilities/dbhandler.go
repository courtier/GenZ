package utilities

import (
	"encoding/json"
	"math/rand"
	"strings"
	"time"

	eggshell "github.com/courtier/eggshell"
	"github.com/google/uuid"
	"github.com/gookit/color"
)

//SavedEmbed object, holds info about an embed generate message
type SavedEmbed struct {
	MessageID string
	ChannelID string
}

//Account object, holds info which is 1 line of text, UID and how many times its been reported as not working
type Account struct {
	Info      string
	ID        string
	Reports   int
	Reporters []string
}

//GennedAccount object, holds info about generated account, who generated it, what account, when etc.
type GennedAccount struct {
	UserID    string
	AccountID string
	Time      int64
}

var (
	dbPath string = "accountsdb"
)

//LoadDB loads database
func LoadDB() (db *eggshell.Driver) {

	db, err := eggshell.CreateDriver(dbPath)
	//db, err := eggshell.New(dbPath, nil)
	if err != nil {
		color.Red.Println("error loading database, ", err)
	}

	color.Green.Println("database loaded")

	return

}

//InsertAccounts insert accounts into category in database
func InsertAccounts(db *eggshell.Driver, category string, accountLines []string) (success bool) {
	category = strings.ToLower(category)
	var accounts []interface{}
	for _, accountLine := range accountLines {
		uuid := uuid.New().String()
		account := Account{accountLine, uuid, 0, []string{}}
		accounts = append(accounts, account)
	}
	if err := db.InsertAllDocuments(category, accounts); err != nil {
		return false
	}
	return true
}

//ReceiveAccount receive a random account
func ReceiveAccount(db *eggshell.Driver, category string) Account {

	category = strings.ToLower(category)
	documents, err := db.ReadAll(category)
	if err != nil {
		color.Red.Println("error receiving account, ", err)
	}

	rand.Seed(time.Now().UnixNano())
	account := Account{}
	if len(documents) > 0 {

		accountIndex := rand.Intn(len(documents))

		if err := json.Unmarshal([]byte(documents[accountIndex]), &account); err != nil {
			color.Red.Println("error unmarshaling account, ", err)
		}

	}

	return account

}

//GetAccountFromUUID receive account from uuid
func GetAccountFromUUID(db *eggshell.Driver, category string, accountID string) (account Account) {

	// Read all fish from the database, unmarshaling the response.
	category = strings.ToLower(category)
	if category == "" {
		return
	}
	foundAccount := Account{}
	documents, err := db.ReadFiltered(category, []string{"ID"}, []string{accountID})
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(documents[0]), &foundAccount)

	if err != nil {
		return
	}

	return foundAccount

}

//RemoveCategory remove category from db
func RemoveCategory(db *eggshell.Driver, category string) bool {

	// Read all fish from the database, unmarshaling the response.
	category = strings.ToLower(category)
	if category == "" {
		return false
	}

	err := db.DeleteCollection(category)
	if err != nil {
		return false
	}

	return true

}

//ReportAccount reports account as not working
func ReportAccount(db *eggshell.Driver, category string, accountID string, reporterID string, maxReports int) (worked bool) {

	category = strings.ToLower(category)
	if category == "" {
		return false
	}

	foundAccount := Account{}
	documents, err := db.ReadFiltered(category, []string{"ID"}, []string{accountID})
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(documents[0]), &foundAccount)

	if err != nil {
		return
	}

	reporters := foundAccount.Reporters
	reporters = append(reporters, reporterID)
	reportAmount := foundAccount.Reports + 1
	if reportAmount == maxReports {
		if err := db.DeleteFiltered(category, []string{"ID"}, []string{accountID}); err != nil {
			return false
		}
		return true
	}

	account := Account{foundAccount.Info, foundAccount.ID, reportAmount, reporters}
	if err := db.InsertDocument(category, account); err != nil {
		return false
	}

	return true

}

//SaveEmbed insert embed message into database
func SaveEmbed(db *eggshell.Driver, messageID string, channelID string) (success bool) {
	savedEmbed := SavedEmbed{messageID, channelID}
	if err := db.InsertDocument("embed123456789", savedEmbed); err != nil {
		return false
	}
	return true
}

//RetrieveEmbed retrieve embed message from database
func RetrieveEmbed(db *eggshell.Driver) (savedEmbed SavedEmbed) {
	foundEmbed := SavedEmbed{}
	documents, err := db.ReadAll("embed123456789")
	if err != nil {
		color.Red.Println("error retrieving embed, ", err)
	}

	if len(documents) > 0 {

		if err := json.Unmarshal([]byte(documents[0]), &foundEmbed); err != nil {
			color.Red.Println("error unmarshaling embed, ", err)
		}

	}

	return foundEmbed
}

//DeleteEmbed delete embed message from database
func DeleteEmbed(db *eggshell.Driver) {
	db.DeleteCollection("embed123456789")
}

//SaveGennedAccount insert when and what account someone genned
func SaveGennedAccount(db *eggshell.Driver, gennedAccount GennedAccount) (success bool) {
	if err := db.InsertDocument("gennedaccounts", gennedAccount); err != nil {
		return false
	}
	return true
}

//DeleteGennedAccount delete when and what account someone genned
func DeleteGennedAccount(db *eggshell.Driver, userID string) (success bool) {
	if err := db.DeleteFiltered("gennedaccounts", []string{"UserID"}, []string{userID}); err != nil {
		return false
	}
	return true
}

//RetrieveGennedAccount retrieve genned account from database
func RetrieveGennedAccount(db *eggshell.Driver, userID string) (genned GennedAccount, err error) {
	foundAccount := GennedAccount{}
	documents, err := db.ReadFiltered("gennedaccounts", []string{"UserID"}, []string{userID})
	if err != nil {
		color.Red.Println("error retrieving genned account, ", err)
		return foundAccount, err
	}

	if len(documents) > 0 {

		if err := json.Unmarshal([]byte(documents[0]), &foundAccount); err != nil {
			color.Red.Println("error unmarshaling genned account, ", err)
			return foundAccount, err
		}

	}

	return foundAccount, nil
}

//GetAllCategories lists all categories
func GetAllCategories(db *eggshell.Driver) []string {
	categories := db.GetAllCollections()
	cleanCategories := []string{}
	for _, ele := range categories {
		if ele != "embed123456789" && ele != "gennedaccounts" {
			cleanCategories = append(cleanCategories, ele)
		}
	}
	return cleanCategories
}

func countAccounts(db *eggshell.Driver, category string) int {
	category = strings.ToLower(category)
	if category == "" {
		return 0
	}
	records, err := db.ReadAll(category)
	if err != nil {
		return 0
	}
	return len(records)
}
