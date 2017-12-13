package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	Token   string
	BotID   string
	AdminID string
)

const (
	GenderNeuter byte = iota
	GenderMale
	GenderFemale
)

var Spouse = [...]string{"spouse", "hazubando", "waifu"}
var Child = [...]string{"child", "son", "daughter"}
var Gender = [...]string{"enby", "male", "female"}
var pa = [...]string{"theirs", "his", "hers"}
var ps = [...]string{"they", "he", "she"}
var po = [...]string{"them", "him", "her"}
var pp = [...]string{"their", "his", "her"}
var pr = [...]string{"themself", "himself", "herself"}

type Human interface {
	GetName() string
	GetGender() byte
}

type BotWaifu struct {
	Name    string
	Gender  byte
	Picture string
}

func (b *BotWaifu) GetName() string { return b.Name }
func (b *BotWaifu) GetGender() byte { return b.Gender }

type BotUser struct {
	Nickname string
	Gender   byte
	Waifus   []*BotWaifu
	Children []*BotWaifu
}

func (b *BotUser) GetName() string { return b.Nickname }
func (b *BotUser) GetGender() byte { return b.Gender }

type BotState struct {
	Users         map[string]*BotUser
	CommandPrefix string
}

type BotCmd func(*discordgo.Session, *discordgo.MessageCreate)

var Global BotState

var Commands map[string]BotCmd
var Usages map[string]string
var Comforts []string

func reply(s *discordgo.Session, m *discordgo.MessageCreate, msg string) {
	fmt.Printf("me -> %s: %s\n", m.ChannelID, msg)
	_, _ = s.ChannelMessageSend(m.ChannelID, msg)
}

func adduserifne(m *discordgo.MessageCreate) {
	if Global.Users[m.Author.ID] == nil {
		ret := new(BotUser)
		ret.Nickname = m.Author.Username
		Global.Users[m.Author.ID] = ret
	}
}

func fetchWaifu(u *BotUser) *BotWaifu {
	if u.Waifus == nil {
		return nil
	} else if len(u.Waifus) == 0 {
		return nil
	} else {
		return u.Waifus[0]
	}
}

func getGender(s *discordgo.Session, m *discordgo.MessageCreate) {
	var id string
	var u *BotUser
	if len(m.Mentions) > 0 {
		id = m.Mentions[0].ID
	} else {
		id = m.Author.ID
	}
	u = Global.Users[id]
	if u == nil {
		reply(s, m, "I've no idea who that is!")
	} else {
		gen := u.Gender
		reply(s, m, fmt.Sprintf("%s's gender is %s (%s, %s)", u.Nickname, Gender[gen],
			ps[gen], po[gen]))
	}
}

func getSpouseString(u *BotUser) string {
	wifu := fetchWaifu(u)
	ret := ""
	if wifu == nil {
		ret = fmt.Sprintf("Looks like %s doesn't have a waifu...\n", u.Nickname)
	} else if len(u.Waifus) == 1 {
		pic := ""
		if wifu.Picture != "" {
			pic = " (" + wifu.Picture + ")"
		}
		ret = fmt.Sprintf(
			"According to the databanks, %s's %s is %s%s\n",
			u.Nickname, Spouse[wifu.Gender], wifu.Name, pic)
	} else {
		ret = fmt.Sprintf("%s has %d spouses:\n", u.Nickname, len(u.Waifus))
		for i, waifu := range u.Waifus {
			pic := ""
			if waifu.Picture != "" {
				pic = " (" + waifu.Picture + ")"
			}
			ret += fmt.Sprintf(
				"%d) %s %s, %s%s\n", i+1,
				pp[u.Gender], Spouse[waifu.Gender], waifu.Name, pic)
		}
	}
	return ret
}

func getWaifu(s *discordgo.Session, m *discordgo.MessageCreate) {
	var id string
	var u *BotUser
	if len(m.Mentions) > 0 {
		id = m.Mentions[0].ID
	} else {
		id = m.Author.ID
	}
	u = Global.Users[id]
	if u == nil {
		reply(s, m, "I've no idea who that is!")
	} else {
		reply(s, m, getSpouseString(u))
	}
}

func getFamily(s *discordgo.Session, m *discordgo.MessageCreate) {
	var id string
	var u *BotUser
	if len(m.Mentions) > 0 {
		id = m.Mentions[0].ID
	} else {
		id = m.Author.ID
	}
	u = Global.Users[id]
	if u == nil {
		reply(s, m, "I've no idea who that is!")
	} else {
		ret := getSpouseString(u)
		if u.Children == nil {
			ret += fmt.Sprintf("Looks like %s doesn't have any children...", u.Nickname)
		} else if len(u.Children) == 0 {
			ret += fmt.Sprintf("Looks like %s doesn't have any children...", u.Nickname)
		} else {
			ret += fmt.Sprintf("%s's children are:", u.Nickname)
			for _, child := range u.Children {
				pic := ""
				if child.Picture != "" {
					pic = "(" + child.Picture + ")"
				}
				ret += fmt.Sprintf(
					"\n%s %s, %s %s",
					pp[u.Gender], Child[child.Gender], child.Name, pic)
			}
		}
		reply(s, m, ret)
	}
}

func pronouns(user Human, waifu Human, str string) string {
	ug := user.GetGender()
	wg := waifu.GetGender()
	ret := str
	ret = strings.Replace(ret, "%a", pa[ug], -1)
	ret = strings.Replace(ret, "%wa", pa[wg], -1)
	ret = strings.Replace(ret, "%s", ps[ug], -1)
	ret = strings.Replace(ret, "%ws", ps[wg], -1)
	ret = strings.Replace(ret, "%o", po[ug], -1)
	ret = strings.Replace(ret, "%wo", po[wg], -1)
	ret = strings.Replace(ret, "%p", pp[ug], -1)
	ret = strings.Replace(ret, "%wp", pp[wg], -1)
	ret = strings.Replace(ret, "%r", pr[ug], -1)
	ret = strings.Replace(ret, "%wr", pr[wg], -1)
	ret = strings.Replace(ret, "%n", user.GetName(), -1)
	ret = strings.Replace(ret, "%wn", waifu.GetName(), -1)
	return ret
}

func nickname(s *discordgo.Session, m *discordgo.MessageCreate) {
	adduserifne(m)
	words := strings.Split(m.Content, " ")
	u := Global.Users[m.Author.ID]
	if len(words) > 1 {
		newnick := strings.Join(words[1:], " ")
		reply(s, m, fmt.Sprintf("Setting %s's nickname to %s", u.Nickname, newnick))
		u.Nickname = newnick
	} else {
		reply(s, m, fmt.Sprintf("Your nickname is %s", u.Nickname))
	}
}

func setGender(s *discordgo.Session, m *discordgo.MessageCreate) {
	adduserifne(m)
	words := strings.Split(m.Content, " ")
	u := Global.Users[m.Author.ID]
	if len(words) > 1 {
		gen := GenderNeuter
		if strings.HasPrefix(strings.ToLower(words[1]), "f") {
			gen = GenderFemale
		}
		if strings.HasPrefix(strings.ToLower(words[1]), "m") {
			gen = GenderMale
		}
		u.Gender = gen
		reply(s, m, fmt.Sprintf("Setting %s's gender to %s", u.Nickname, Gender[gen]))
	} else {
		reply(s, m, fmt.Sprintf("%s's gender is %s", u.Nickname, Gender[u.Gender]))
	}
}

func comfort(s *discordgo.Session, m *discordgo.MessageCreate) {
	comfortUser(s, m, false)
}

func rcomfort(s *discordgo.Session, m *discordgo.MessageCreate) {
	comfortUser(s, m, true)
}

func comfortUser(s *discordgo.Session, m *discordgo.MessageCreate, rev bool) {
	var id string
	var name string
	var u *BotUser
	if len(m.Mentions) > 0 {
		id = m.Mentions[0].ID
		name = m.Mentions[0].Username
	} else {
		id = m.Author.ID
		name = m.Author.Username
	}
	u = Global.Users[id]
	if u == nil {
		reply(s, m, fmt.Sprintf("_cuddles %s close_", name))
	} else {
		name = u.Nickname
		wifu := fetchRandWaifu(u)
		if wifu == nil {
			reply(s, m, fmt.Sprintf("_cuddles %s close_", name))
		} else {
			if rev {
				reply(s, m, pronouns(wifu, u, randoms(Comforts)))
			} else {
				reply(s, m, pronouns(u, wifu, randoms(Comforts)))
			}
		}
	}
}

func waifuDel(s *discordgo.Session, m *discordgo.MessageCreate) {
	adduserifne(m)
	words := strings.Split(m.Content, " ")
	gen := GenderFemale
	if strings.Contains(strings.ToLower(words[0]), "husbando") {
		gen = GenderMale
	}
	if strings.Contains(strings.ToLower(words[0]), "spouse") {
		gen = GenderNeuter
	}
	if len(words) > 1 {
		var wname string = strings.Join(words[1:], " ")
		if Global.Users[m.Author.ID].Waifus == nil {
			reply(s, m, "But you don't have a waifu!")
		} else {
			u := Global.Users[m.Author.ID]
			for i, waifu := range u.Waifus {
				if waifu.Name == wname && waifu.Gender == gen {
					reply(s, m, fmt.Sprintf("Removing %s from %s's waifus",
						wname, m.Author.Username))
					copy(u.Waifus[i:], u.Waifus[i+1:])
					u.Waifus[len(u.Waifus)-1] = nil // or the zero value of T
					u.Waifus = u.Waifus[:len(u.Waifus)-1]
					return
				}
			}
			reply(s, m, "Couldn't find that waifu in your waifu list!")
		}
	}
}

func childDel(s *discordgo.Session, m *discordgo.MessageCreate) {
	adduserifne(m)
	words := strings.Split(m.Content, " ")
	gen := GenderFemale
	if strings.Contains(strings.ToLower(words[0]), "son") {
		gen = GenderMale
	}
	if strings.Contains(strings.ToLower(words[0]), "child") {
		gen = GenderNeuter
	}
	if len(words) > 1 {
		var wname string = strings.Join(words[1:], " ")
		if Global.Users[m.Author.ID].Children == nil {
			reply(s, m, "But you don't have any children!")
		} else {
			u := Global.Users[m.Author.ID]
			for i, child := range u.Children {
				if child.Name == wname && child.Gender == gen {
					reply(s, m, fmt.Sprintf("Removing %s from %s's children",
						wname, m.Author.Username))
					copy(u.Children[i:], u.Children[i+1:])
					u.Children[len(u.Children)-1] = nil // or the zero value of T
					u.Children = u.Children[:len(u.Children)-1]
					return
				}
			}
			reply(s, m, fmt.Sprintf("%s is not one of your children!", wname))
		}
	}
}

func waifuPicAdd(s *discordgo.Session, m *discordgo.MessageCreate) {
	adduserifne(m)
	words := strings.Split(m.Content, " ")
	if len(words) > 2 {
		var pic string = words[1]
		if pic == "" {
			reply(s, m, "Please provide a picture")
			return
		}
		var wname string = strings.Join(words[2:], " ")
		if Global.Users[m.Author.ID].Waifus != nil {
			u := Global.Users[m.Author.ID]
			for _, waifu := range u.Waifus {
				if waifu.Name == wname {
					reply(s, m, fmt.Sprintf("Adding a picture of %s - %s",
						wname, pic))
					waifu.Picture = pic
					return
				}
			}
		}

		if Global.Users[m.Author.ID].Children != nil {
			u := Global.Users[m.Author.ID]
			for _, c := range u.Children {
				if c.Name == wname {
					reply(s, m, fmt.Sprintf("Adding a picture of %s - %s",
						wname, pic))
					c.Picture = pic
					return
				}
			}
		}
	}
}

func waifuReg(s *discordgo.Session, m *discordgo.MessageCreate) {
	adduserifne(m)
	words := strings.Split(m.Content, " ")
	gen := GenderFemale
	if strings.Contains(strings.ToLower(words[0]), "husbando") {
		gen = GenderMale
	}
	if strings.Contains(strings.ToLower(words[0]), "spouse") {
		gen = GenderNeuter
	}
	spouse := Spouse[gen]
	if len(words) > 1 {
		var wname string = strings.Join(words[1:], " ")
		if Global.Users[m.Author.ID].Waifus == nil {
			Global.Users[m.Author.ID].Waifus = []*BotWaifu{
				&BotWaifu{wname, gen, ""},
			}
		} else {
			Global.Users[m.Author.ID].Waifus = append(Global.Users[m.Author.ID].Waifus,
				&BotWaifu{wname, gen, ""})
		}
		reply(s, m, fmt.Sprintf("Setting %s's %s to %s",
			m.Author.Username, spouse, wname))
		fmt.Println(m.Author.ID, spouse, wname)
	}
}

func addChild(s *discordgo.Session, m *discordgo.MessageCreate) {
	adduserifne(m)
	words := strings.Split(m.Content, " ")
	gen := GenderFemale
	if strings.Contains(strings.ToLower(words[0]), "son") {
		gen = GenderMale
	}
	if strings.Contains(strings.ToLower(words[0]), "child") {
		gen = GenderNeuter
	}
	child := Child[gen]
	if len(words) > 1 {
		var wname string = strings.Join(words[1:], " ")
		if Global.Users[m.Author.ID].Children == nil {
			Global.Users[m.Author.ID].Children = []*BotWaifu{
				&BotWaifu{wname, gen, ""},
			}
		} else {
			Global.Users[m.Author.ID].Children = append(
				Global.Users[m.Author.ID].Children, &BotWaifu{wname, gen, ""})
		}
		reply(s, m, fmt.Sprintf("Setting %s's %s to %s",
			m.Author.Username, child, wname))
		fmt.Println(m.Author.ID, child, wname)
	}
}

func help(s *discordgo.Session, m *discordgo.MessageCreate) {
	words := strings.Split(m.Content, " ")
	if len(words) > 1 {
		cmd := strings.TrimPrefix(strings.Join(words[1:], " "), Global.CommandPrefix)
		if Usages[cmd] == "" {
			reply(s, m, fmt.Sprintf("The help system doesn't know about %s%s",
				Global.CommandPrefix, cmd))
		} else {
			reply(s, m, fmt.Sprintf("%s%s - %s", Global.CommandPrefix,
				cmd, Usages[cmd]))
		}
	} else {
		rep := "tewibot - a spiritual successor to the lainbot family of irc bots.\nSupported commands (type &help _command_ for usage text):\n"
		for key, _ := range Commands {
			rep += Global.CommandPrefix + key + ", "
		}
		reply(s, m, rep)
	}
}

func adminInfo(s *discordgo.Session, m *discordgo.MessageCreate) {
	rep := ""
	if AdminID == "" {
		rep = "There is no admin set for the bot."
	} else if isSenderAdmin(m) {
		rep = "You are the admin."
	} else {
		admin, err := s.User(AdminID)
		if err == nil {
			rep = fmt.Sprintf("%s is the admin.", admin.Mention())
		} else {
			rep = fmt.Sprintf("%s is the admin, but I can't mention them...", AdminID)
		}
	}
	reply(s, m, rep)
}

func isSenderAdmin(m *discordgo.MessageCreate) bool {
	return AdminID != "" && m.Author.ID == AdminID
}

func addCommand(c BotCmd, usage string, aliases ...string) {
	for _, alias := range aliases {
		Commands[alias] = c
		Usages[alias] = usage
	}
}

func init() {
	Commands = make(map[string]BotCmd)
	Usages = make(map[string]string)
	addCommand(waifuReg, "Register your waifu with the bot", "waifureg", "husbandoreg", "setwaifu", "sethusbando", "spousereg", "setspouse")
	addCommand(waifuDel, "Delete a previously registered waifu", "waifudel", "husbandodel", "spousedel")
	addCommand(childDel, "Delete a previously registered child", "daughterdel", "sondel", "childdel")
	addCommand(getGender, "Print your (or someone else's) gender", "gender", "getgender")
	addCommand(getWaifu, "Print your (or someone else's) waifu", "waifu", "husbando", "spouse")
	addCommand(comfort, "Dispense hugs and other niceness from your waifu", "comfort", "hug")
	addCommand(rcomfort, "Dispense hugs and other niceness to your waifu", "rcomfort", "rhug")
	addCommand(setGender, "Set your gender - m, f, x\nThis affects which pronouns the bot will use for you (he, she, they)", "setgender", "genderreg")
	addCommand(addChild, "Register one of your children with the bot", "setchild", "childreg", "setdaughteru", "daughterureg", "setsonfu", "sonfureg")
	addCommand(getFamily, "Print your (or someone else's) family", "family", "getfamily")
	addCommand(nickname, "If given a nickname, set your nickname to that. Otherwise, print your nickname.", "nick", "nickname", "setnick", "setnickname")
	addCommand(help, "Access the on-line help system", "help", "usage", "sos")
	addCommand(adminInfo, "Print information about the admin", "admin")
	addCommand(waifuPicAdd, "Add a picture to your waifu; e.g. &picadd http://i.imgur.com/Gqf1rGi.jpg Miku", "picadd")
	InitGlobal()
	InitComforts()

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&AdminID, "a", "", "Admin's Discord ID")
	flag.Parse()
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Get the account information.
	u, err := dg.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}

	// Store the account ID for later use.
	BotID = u.ID

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sigchan := make(chan os.Signal, 10)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan
	fmt.Println("Recieved interrupt, exiting gracefully")
	SaveGlobal()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == BotID {
		return
	}

	fmt.Printf("%s (%s) -> %s: %s\n", m.Author.Username, m.Author.ID, m.ChannelID, m.Content)

	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		reply(s, m, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		reply(s, m, "Ping!")
	}

	// Informal standard
	if m.Content == ".bots" {
		reply(s, m, "tewibot reporting in! [Golang] https://github.com/japanoise/tewibot")
	}

	if len(m.Content) > len(Global.CommandPrefix) {
		if strings.HasPrefix(m.Content, Global.CommandPrefix) {
			run := Commands[strings.ToLower(strings.Split(strings.TrimPrefix(
				m.Content, Global.CommandPrefix), " ")[0])]
			if run != nil {
				run(s, m)
			}
		}
	}
}
