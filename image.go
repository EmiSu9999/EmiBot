package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

var lock sync.Mutex

func fetchImageDanbooru(tag string) string {
	if Global.DanbooruLogin == "" || Global.DanbooruAPIKey == "" {
		return `Error: Can't do API requests without both a Danbooru Login & API key. https://danbooru.donmai.us/wiki_pages/43568`
	}
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	// Generated by curl-to-Go: https://mholt.github.io/curl-to-go
	body := strings.NewReader(`tags=` + tag + ` rating:safe&limit=1&random=true`)
	req, err := http.NewRequest("GET", "https://danbooru.donmai.us/posts.json", body)
	if err != nil {
		return err.Error()
	}

	req.SetBasicAuth(Global.DanbooruLogin, Global.DanbooruAPIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := netClient.Do(req)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err.Error()
		}
		if len(body) < 3 {
			return "Nobody here but us chickens!"
		}
		return imageLinkForJson(body)
	} else {
		return resp.Status
	}
}

// PLEASE DO NOT USE THIS FUNCTION ANYWHERE ELSE, IT'S TERRIBLE
// Assumes:
// - if it's passed a list, it's a singleton list NOT AN EMPTY LIST
// - json data isn't empty
// These are huge assumptions. Don't blame me if you pass json that you've not checked these against into this function.
// If you really want to do something like what the bot does, tweak the method above.
func imageLinkForJson(b []byte) string {
	// Very hacky. I cba to do the whole danbooru API, so we're just throwing this in as an unstructured JSON object.
	if b[0] == '[' {
		// xtreme hack: unlistify it. Makes the rest a little easier.
		// We assume it's a singleton list.
		b[0] = ' '
		b[len(b)-1] = ' '
	}
	var obj interface{}
	err := json.Unmarshal(b, &obj)
	if err != nil {
		return err.Error()
	}
	objmap := obj.(map[string]interface{})
	fileurl := objmap["file_url"]
	if fileurl == nil {
		return "Malformed json data (nothing for file_url found)"
	}
	switch fut := fileurl.(type) {
	case string:
		if strings.HasPrefix(fut, "http") {
			return strings.Replace(strings.Replace(fut, "https//", "https://", 1), "http//", "http://", 1)
		} else {
			return "https://danbooru.donmai.us" + fut
		}
	default:
		return "Malformed json data (wrong type for file_url; was expecting string)"
	}
}

func watchReactions(session *discordgo.Session, message *discordgo.Message, requester *discordgo.User, msg string) {
	// GO Routine
	go func() {
		for i := 0; i < 31; i++ { // Watch Message for 30 seconds
			users, _ := session.MessageReactions(message.ChannelID, message.ID, "🚫", 100, "", "")
			for _, user := range users {
				if user.ID == requester.ID {
					session.ChannelMessageDelete(message.ChannelID, message.ID)
					promptBlacklist(session, message, requester, msg) // After deleting message, ask user if they want to blacklist it
					return
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

func promptBlacklist(session *discordgo.Session, message *discordgo.Message, requester *discordgo.User, msg string) {
	prompt := requester.Mention() + " React with 👍 if you wish to blacklist the deleted image"
	deletePrompt, error := session.ChannelMessageSend(message.ChannelID, prompt)
	if error == nil {
		session.MessageReactionAdd(deletePrompt.ChannelID, deletePrompt.ID, "👍")
		for j := 0; j < 31; j++ {
			users, _ := session.MessageReactions(deletePrompt.ChannelID, deletePrompt.ID, "👍", 100, "", "")
			for _, user := range users {
				if user.ID == requester.ID {
					addToBlacklist(requester, msg)
					session.ChannelMessageSend(message.ChannelID, "Successfully added to blacklist.")
					return
				}
			}
		}
	}
}

func addToBlacklist(user *discordgo.User, msg string) {
	lock.Lock()
	if val, ok := Blacklist[user.ID]; ok {
		Blacklist[user.ID] = append(val, msg)
	} else {
		Blacklist[user.ID] = []string{msg}
	}
	lock.Unlock()
}

func checkBlacklist(user *discordgo.User, msg string) bool {
	lock.Lock()
	if val, ok := Blacklist[user.ID]; ok {
		for _, entry := range Blacklist[user.ID] {
			if entry == msg {
				lock.Unlock()
				return true
			}
		}
		Blacklist[user.ID] = append(val, msg)
	}
	lock.Unlock()
	return false
}
