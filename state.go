package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func InitGlobal() {
	f, err := os.Open("waifus.json")
	if err == nil {
		dec := json.NewDecoder(f)
		if err = dec.Decode(&Global); err != nil {
			fmt.Println(err.Error(), ", using a blank db for now.")
			Global = BotState{make(map[string]*BotUser), "&", "", ""}
		}
	} else {
		fmt.Println(err.Error(), ", using a blank db for now.")
		Global = BotState{make(map[string]*BotUser), "&", "", ""}
	}
}

func LoadComfortsList(filename string, list interface{}) error {
	f, err := os.Open(filename)
	if err == nil {
		dec := json.NewDecoder(f)
		err = dec.Decode(list)
	}
	return err
}

func InitComforts() {
	err := LoadComfortsList("comforts.json", &Comforts)
	if err != nil {
		fmt.Println(err.Error(), ", using minimal comforts db for now.")
		Comforts = []string{"_%wn hugs %n_"}
	}
	err = LoadComfortsList("childcomforts.json", &ChildComforts)
	if err != nil {
		fmt.Println(err.Error(), ", using minimal child comforts db for now.")
		ChildComforts = []string{"_%wn hugs %n_"}
	}
	err = LoadComfortsList("childrcomforts.json", &ChildReverseComforts)
	if err != nil {
		fmt.Println(err.Error(), ", using minimal child reverse comforts db for now.")
		ChildReverseComforts = []string{"_%wn hugs %n_"}
	}
}

// Loads custom_responses.json and parses custom responses to given user input for later use
func InitCustomResponses() {
	// Filenames could be pulled out into a config file in the long run...
	jsonFile, err := os.Open("custom_responses.json")

	// Error handling in case file does not exist
	if err != nil {
		fmt.Println(err, ", no custom responses were loaded. custom_responses.json might not exist.")
		return
	}

	// Close jsonFile as soon as this function returns
	defer jsonFile.Close()

	// Unmarshal can't work with a file handle, and needs the bytes.
	// Errors are ignored here, as there is no way they should happen, as the existence of the file has already been tested for
	jsonBytes, _ := ioutil.ReadAll(jsonFile)

	// create a temporary map, to hold the keys before making lowercase
	temporaryMap := make(map[string]string)

	// Unmarshal into CustomResponses map
	err = json.Unmarshal(jsonBytes, &temporaryMap)

	// Error handling in case json is invalid
	if err != nil {
		fmt.Println(err, ", no custom responses were loaded. custom_responses.json might be malformed.")
		return
	}

	// Make everything lowercase
	for k, v := range temporaryMap {
		CustomResponses[strings.Title(k)] = v
	}
}

func SaveGlobal() {
	f, err := os.Create("waifus.json")
	if err == nil {
		defer f.Close()
		data, err := json.MarshalIndent(&Global, "", "\t")
		if err != nil {
			fmt.Println(err.Error())
		} else {
			f.Write(data)
		}
	} else {
		fmt.Println(err.Error())
	}
}
