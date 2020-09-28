package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	// Used for file watching
	"github.com/fsnotify/fsnotify"
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
		f.Close()
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

// For later reloading of facts
func LoadFacts(filename string, list interface{}) error {
	f, err := os.Open(filename)
	if err == nil {
		dec := json.NewDecoder(f)
		err = dec.Decode(list)
		f.Close()
	}

	return err
}

// Load facts from facts.json into Facts array. Add placeholder if file can not be loaded.
func InitFacts() {
	err := LoadFacts("facts.json", &Facts)
	if err != nil {
		fmt.Println(err.Error(), ", Added placeholder fact")
		Facts = []string{"Facts were not loaded for some reason... but here's one: I love Emilia!"}
	}
}

// Loads custom_responses.json and parses custom responses to given user input for later use
func InitCustomResponses() error {
	// Filenames could be pulled out into a config file in the long run...
	jsonFile, err := os.Open("custom_responses.json")

	// Error handling in case file does not exist
	if err != nil {
		fmt.Println(err, ", no custom responses were loaded. custom_responses.json might not exist.")
		return err
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
		return err
	}

	CustomResponses = make(map[string]string)
	// Make everything lowercase
	for k, v := range temporaryMap {
		CustomResponses[strings.Title(k)] = v
	}

	return nil
}

// Decides what to do, when certain files in the directory are altered
func handleReload(filename string) {
	var err error = nil      // Used for error handling/ notifying of errors
	var changed bool = false // Used to check if the file altered was among the important ones

	// Windows filewatcher sets "xy.json" as name, while Linux uses "./xy.json". This solves that.
	filename = strings.ReplaceAll(filename, "./", "")

	switch filename {

	case "custom_responses.json":
		fmt.Printf("Reloading %s. If file is malformed, old set will be kept...\n", filename)
		err = InitCustomResponses()
		changed = true
		break

	case "comforts.json":
		fmt.Printf("Reloading %s. If file is malformed, old set will be kept...\n", filename)
		err = LoadComfortsList(filename, &Comforts)
		changed = true
		break

	case "childcomforts.json":
		fmt.Printf("Reloading %s. If file is malformed, old set will be kept...\n", filename)
		err = LoadComfortsList(filename, &ChildComforts)
		changed = true
		break

	case "childrcomforts.json":
		fmt.Printf("Reloading %s. If file is malformed, old set will be kept...\n", filename)
		err = LoadComfortsList(filename, &ChildReverseComforts)
		changed = true
		break
	case "facts.json":
		fmt.Printf("Reloading %s. If file is malformed, old set will be kept...\n", filename)
		err = LoadFacts(filename, &Facts)
		changed = true
		break

	}

	if err != nil {
		fmt.Println(err, " , file couldn't be reloaded. Staying with old contents...")
	} else if err == nil && changed {
		// Only print this when changed has been set true, so nothing is printed when trivial file changes occur (eg. waifus.json)
		fmt.Println("Successfully reloaded ", filename)
	}
}

// Attach a file watcher, so changes in configuration files can be reloaded on the go.. (no pun intended)
func AttachWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println(err, ", couldn't attach file watcher. Bot has to be restarted in case of changes..")
		return
	}

	// Go routine. As such, this piece of code will run concurrently to the rest of the code as long as the bot is up
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// "if the even that happened is a write to a file"
				if event.Op&fsnotify.Write == fsnotify.Write {
					handleReload(event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				// Shouldn't happen, but just in case
				fmt.Println("File Watcher Error:", err)
			}
		}
	}()

	//Attach the watcher to the working directory
	err = watcher.Add("./")
	if err != nil {
		fmt.Println(err, ", couldn't attach file watcher. Bot has to be restarted in case of changes..")
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
