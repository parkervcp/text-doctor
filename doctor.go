package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

var (
	// DocConfig is the config for the text-doctor
	DocConfig   docConfig
	currentFile sheetResponse
)

type docConfig struct {
	Sheet struct {
		ID              string `json:"id"`
		Table           string `json:"table"`
		CellsStart      string `json:"cells_start"`
		CellEnd         string `json:"cell_end"`
		Columns         []int  `json:"columns"`
		RefreshInterval int    `json:"refresh_interval"`
	} `json:"sheet"`
	File struct {
		Location       string `json:"location"`
		UpdateInterval int    `json:"update_interval"`
		Format         string `json:"format"`
	} `json:"file"`
}

type sheetResponse struct {
	Responses map[int][]string
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func writeFile(line string) {
	fo, err := os.Create(DocConfig.File.Location)
	if err != nil {
		log.Fatalf("Could not write to file")
	}
	defer fo.Close()

	fo.WriteString(line)

	fo.Close()
}

func main() {

	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// Repeat this every XX seconds until the application is closed.
	for {
		// Gets the spreadsheet information
		// https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
		log.Printf("Getting info from spreadsheet")
		readRange := DocConfig.Sheet.Table + "!" + DocConfig.Sheet.CellsStart + ":" + DocConfig.Sheet.CellEnd
		resp, err := srv.Spreadsheets.Values.Get(DocConfig.Sheet.ID, readRange).Do()
		if err != nil {
			log.Fatalf("Unable to retrieve data from sheet: %v", err)
		}

		// write data to a map
		rows := make(map[int][]string)

		for ID, row := range resp.Values {
			// resets the value of columns every loop.
			var columns []string
			// add all the comlumn data to a string array
			for _, column := range DocConfig.Sheet.Columns {
				columns = append(columns, row[column].(string))
			}
			// write string array to map under ID for the
			rows[ID] = columns
		}

		//compare stored data with the new data and update if needed.
		if !reflect.DeepEqual(rows, currentFile.Responses) {
			var lines []string

			log.Printf("Change in the spreadsheet. Updating the file.")
			currentFile.Responses = rows

			// writing all the things to a string to make lines pretty and format them.
			for _, row := range rows {
				format := DocConfig.File.Format
				for _, column := range DocConfig.Sheet.Columns {
					format = strings.Replace(format, "&"+strconv.Itoa(column)+"&", row[column], 1)
				}
				lines = append(lines, format)
			}

			writeFile(strings.Join(lines, ""))
		} else {
			log.Printf("Spreadsheet has not updated.")
		}

		log.Printf("Sleeping %v seconds.", DocConfig.Sheet.RefreshInterval)

		// sleep for a minimum of 60 seconds before querying the API again
		time.Sleep(time.Duration(DocConfig.Sheet.RefreshInterval) * time.Second)
	}
}

func init() {
	//log.SetOutput(os.Stdout)
	// Open our jsonFile
	jsonFile, err := os.Open("config.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		log.Fatalf("Error loading config.")
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &DocConfig)

	if DocConfig.Sheet.ID == "" {
		log.Fatalf("No Sheet ID in the config.")
	}

	if DocConfig.Sheet.CellsStart == "" || DocConfig.Sheet.CellEnd == "" {
		log.Fatalf("A starting and ending cell is required.")
	}

	if len(DocConfig.Sheet.Columns) < 1 {
		log.Fatalf("At least one column must be set for values")
	}

	if DocConfig.Sheet.RefreshInterval < 60 {
		DocConfig.Sheet.RefreshInterval = 60
	}
}
