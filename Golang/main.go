// author: mwcsur mwcsur@gmail.com

package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type data_model_edge struct {
	Channel                int    `json:"Channel"`
	DsatDefinitionsVersion int    `json:"DsatDefinitionsVersion"`
	DsatEventTimestamps    string `json:"DsatEventTimestamps"`
	DsatEventTypeIds       string `json:"DsatEventTypeIds"`
	DsatEventValues        string `json:"DsatEventValues"`
	DsatUrlToEventsMapping string `json:"DsatUrlToEventsMapping"`
	DsatUrls               string `json:"DsatUrls"`
	EventInfoLevel         int    `json:"EventInfo.Level"`
	AppVersion             string `json:"app_version"`
	ClientId               int    `json:"client_id"`
	IsDmaApplicable        int    `json:"isDmaApplicable"`
	PayloadId              int    `json:"payload_id"`
	PopSample              int    `json:"pop_sample"`
	UtcFlags               int    `json:"utc_flags"`
}

type data_model_app struct {
	Name        string `json:"Name"`
	Publisher   string `json:"Publisher"`
	Version     string `json:"Version"`
	RootDirPath string `json:"RootDirPath"`
	InstallDate string `json:"InstallDate"`
}

type data_model_app_exec struct {
	Version string `json:"AppVersion"`
	AppId   string `json:"AppId"`
}

type data_model_user_default struct {
	DefaultApp           string `json:"DefaultApp"`
	DefaultBrowserProgId string `json:"DefaultBrowserProgId"`
}

type query_data_edge struct {
	Ver  string          `json:"ver"`
	Name string          `json:"name"`
	Time string          `json:"time"`
	Key  string          `json:"iKey"`
	Data data_model_edge `json:"data"`
}

type query_data_app struct {
	Ver  string         `json:"ver"`
	Name string         `json:"name"`
	Time string         `json:"time"`
	Key  string         `json:"iKey"`
	Data data_model_app `json:"data"`
}

type query_data_app_exec struct {
	Data data_model_app_exec `json:"data"`
}

type query_data_user_default struct {
	Ver  string                  `json:"ver"`
	Name string                  `json:"name"`
	Time string                  `json:"time"`
	Key  string                  `json:"iKey"`
	Data data_model_user_default `json:"data"`
}

func output_data(entries [][]string, output string, category string) {
	var csv_f, err = os.Create(fmt.Sprintf("%s/%s.csv", output, category))
	if err != nil {
		log.Fatalln(err)
	}

	var writer = csv.NewWriter(csv_f)

	for _, v := range entries {
		we := writer.Write(v)
		if we != nil {
			log.Println(we)
		}
	}

	writer.Flush()
	if ferr := writer.Error(); ferr != nil {
		log.Println(ferr)
	}

}

func GetEdgeBrowsing(db string, output string) {

	var output_entries [][]string

	var query = `SELECT events_persisted.sid, events_persisted.payload from events_persisted inner join event_tags on events_persisted.full_event_name_hash = event_tags.full_event_name_hash inner join tag_descriptions on event_tags.tag_id = tag_descriptions.tag_id where (tag_descriptions.tag_id = "1" and events_persisted.full_event_name LIKE "%Aria.218d658af29e41b6bc37144bd03f018d.Microsoft.WebBrowser.HistoryJournal%")`
	db_handle, o_err := sql.Open("sqlite3", db)
	if o_err != nil {
		log.Fatalf("Unable to open db file: %s", db)
	}

	defer db_handle.Close()

	log.Printf("[+] Opened db %s\n", db)
	log.Println("[+] Executing Edge Browsing Parsing")

	rows, q_err := db_handle.Query(query)
	if q_err != nil {
		log.Fatalln(q_err)
	}

	//create header slice
	var header = []string{"Visit URLS", "TimeStamps", "SID"}
	output_entries = append(output_entries, header)
	for rows.Next() {
		//var data string

		var raw_sid string
		var raw_myData string
		var e = rows.Scan(&raw_sid, &raw_myData)
		var myData query_data_edge

		var je = json.Unmarshal([]byte(raw_myData), &myData)
		if je != nil {
			log.Println(je)
		}
		if e != nil {
			log.Println(e)
		}

		//build slice
		if myData.Data.DsatUrls != "" {
			var temp_row = []string{myData.Data.DsatUrls, myData.Data.DsatEventTimestamps, raw_sid}
			output_entries = append(output_entries, temp_row)
		}

	}

	//save  to csv
	output_data(output_entries, output, "edge_browsing")
}

func GetApplicationInventory(db string, output string) {
	const query = `SELECT events_persisted.sid, events_persisted.payload from events_persisted inner join event_tags on events_persisted.full_event_name_hash = event_tags.full_event_name_hash inner join tag_descriptions on event_tags.tag_id = tag_descriptions.tag_id where (tag_descriptions.tag_id = 31 and events_persisted.full_event_name="Microsoft.Windows.Inventory.Core.InventoryApplicationAdd")`
	db_handle, err := sql.Open("sqlite3", db)
	if err != nil {
		log.Fatalf("Unable to open db file: %s", db)
	}

	defer db_handle.Close()

	log.Printf("[+] Opened db %s\n", db)
	log.Println("[+] Executing Application Inventory Parsing")

	rows, err := db_handle.Query(query)
	if err != nil {
		log.Fatalln(err)
	}

	var header = []string{"Application Name", "Installation Directory", "Installation Timestamp (UTC)", "Publisher", "Application Version", "SID"}
	var output_entries [][]string
	output_entries = append(output_entries, header)

	for rows.Next() {
		var raw_data string
		var raw_myData string
		var e = rows.Scan(&raw_data, &raw_myData)
		var myData query_data_app

		var je = json.Unmarshal([]byte(raw_myData), &myData)
		if je != nil {
			log.Println(je)
		}
		if e != nil {
			log.Println(e)
		}

		var name = myData.Data.Name
		var root_dir = myData.Data.RootDirPath
		var install_data = myData.Data.InstallDate
		var publisher = myData.Data.Publisher
		var version = myData.Data.Version

		//build slice
		var temp_row = []string{name, root_dir, install_data, publisher, version, raw_data}
		output_entries = append(output_entries, temp_row)

	}

	//save to csv
	output_data(output_entries, output, "application_inventory")

}

func GetApplicationExecution(db string, output string) {
	const query = `SELECT events_persisted.sid, events_persisted.payload from events_persisted inner join event_tags on events_persisted.full_event_name_hash = event_tags.full_event_name_hash inner join tag_descriptions on event_tags.tag_id = tag_descriptions.tag_id where (tag_descriptions.tag_id = 25 and events_persisted.full_event_name="Win32kTraceLogging.AppInteractivitySummary")`
	db_handle, err := sql.Open("sqlite3", db)
	if err != nil {
		log.Fatalf("Unable to open db file: %s", db)
	}

	defer db_handle.Close()

	log.Printf("[+] Opened db %s\n", db)
	log.Println("[+] Executing Application Execution Parsing")

	rows, err := db_handle.Query(query)
	if err != nil {
		log.Fatalln(err)
	}

	var header = []string{"Binary Name", "Execution Timestamp (UTC)", "SHA1 Hash", "Compiler Timestamp (UTC)", "SID"}
	var output_entries [][]string
	output_entries = append(output_entries, header)

	for rows.Next() {
		var raw_data string
		var raw_myData string
		var e = rows.Scan(&raw_data, &raw_myData)

		var myData query_data_app_exec
		var je = json.Unmarshal([]byte(raw_myData), &myData)
		if je != nil {
			log.Println(je)
		}

		if e != nil {
			log.Println(e)
		}

		//all values from app id
		var raw_app_id = myData.Data.AppId
		var raw_app_version = myData.Data.Version

		var binary_data_parts = strings.Split(raw_app_id, "!")
		var app_data_parts = strings.Split(raw_app_version, "!")

		var file_hash = ""
		var time_stamp = ""
		var file_name = ""

		//branch parsing based on prefix
		if binary_data_parts[0][0] == 'W' {
			file_hash = binary_data_parts[1][4:]
			time_stamp = app_data_parts[0]
			file_name = app_data_parts[2]
		} else {
			time_stamp = app_data_parts[0]
			file_name = app_data_parts[3]
		}

		var temp_row = []string{file_name, time_stamp, file_hash, time_stamp, raw_data}
		output_entries = append(output_entries, temp_row)
	}

	output_data(output_entries, output, "application_execution")

}

func GetUserDefaults(db string, output string) {
	const query = `SELECT events_persisted.sid, events_persisted.payload from events_persisted inner join event_tags on events_persisted.full_event_name_hash = event_tags.full_event_name_hash inner join tag_descriptions on event_tags.tag_id = tag_descriptions.tag_id where (tag_descriptions.tag_id = 11 and events_persisted.full_event_name = "Census.Userdefault")`
	db_handle, err := sql.Open("sqlite3", db)
	if err != nil {
		log.Fatalf("Unable to open db file: %s", db)
	}

	defer db_handle.Close()

	log.Printf("[+] Opened db %s\n", db)
	log.Println("[+] Executing User Defaults Parsing")

	rows, err := db_handle.Query(query)
	if err != nil {
		log.Fatalln(err)
	}

	var header = []string{"DefaultBrowser", "DefaultApp", "Time"}
	var output_entries [][]string
	output_entries = append(output_entries, header)

	for rows.Next() {
		var raw_data string
		var raw_myData string
		var e = rows.Scan(&raw_data, &raw_myData)

		if e != nil {
			log.Fatalln(e)
		}

		var myData query_data_user_default
		var je = json.Unmarshal([]byte(raw_myData), &myData)
		if je != nil {
			log.Println(je)
		}

		if e != nil {
			log.Println(e)
		}

		//log.Println(myData)
		var temp_row = []string{myData.Data.DefaultBrowserProgId, myData.Data.DefaultApp, myData.Time}
		output_entries = append(output_entries, temp_row)

	}

	output_data(output_entries, output, "user_defaults")

}

func main_parsing(db string, output string) {
	//check if db exists
	//check if output folder exists or can be made
	if !get_path_exists(db) {
		log.Fatalf("File doesnt exist: %s\n", db)
	}
	if !get_path_exists(output) {
		log.Fatalf("Output location doesnt exist: %s\n", output)
	}

	//edge browsing
	GetEdgeBrowsing(db, output)

	//application inventory
	GetApplicationInventory(db, output)

	//application execution
	GetApplicationExecution(db, output)

	//Get user default
	GetUserDefaults(db, output)

	//Not present in W10 as of 2026 test data
	/*
	   WiFiConnectedEvents
	   SRUMAppActivity
	   WLANScanResults
	   SRUMNetworkUsageActivity
	*/

}

func get_path_exists(entry string) bool {
	_, err := os.Stat(entry)
	if err == nil {
		return true
	} else {
		return false
	}
}

func main() {
	fmt.Println("[*] Loading To Event Transcriber...")
	fmt.Println("[*] Windows Diagnostic Data - EventTranscript.db Parser")

	//get path to db
	var args = os.Args
	if len(args) < 3 {
		log.Println("[-] Invalid args")
		log.Fatalln("[-] usage: <program> <db> <output_location>")
	}

	var db = args[1]
	var output = args[2]

	log.Printf("[+] Loading db: %s\n", db)
	log.Printf("[+] Output locaion: %s\n", output)

	main_parsing(db, output)
}
