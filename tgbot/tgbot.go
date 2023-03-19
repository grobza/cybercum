package tgbot

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dyvdev/cybercum/swatter"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"modernc.org/mathutil"
)

const (
	/*
		список команд для ОтцаБотов:
		add_cum - добавить кума
		enable_semen - включить генерацию сообщений(да/нет)
		enable_phrases - включить фиксированные фразы(да/нет)
		phrase - добавить фразу(через пробел после команды, вернёт номер фразы)
		phrase_remove - убрать фразу по номеру
		ratio - частота сообщений(50 значит, что бот будет писать раз в 50 сообщений)
		length - длина сгенерированных сообщений
	*/
	commandAddCum        = "add_cum"
	commandEnableSemen   = "enable_semen"
	commandEnablePhrases = "enable_phrases"
	commandFixed         = "phrase"
	commandFixedRemove   = "phrase_remove"
	commandRatio         = "ratio"
	commandLength        = "length"

	maxLength = 100
	nefren    = "CAACAgIAAx0CTK3KYQACAQNjDKmYViPp5K-PWxuUKUDpwg0vQQAC9hEAAqx6iEqOhkQYAe2vbSkE"
	dumpTick  = time.Hour
)

type Config struct {
	BotId               string   // айди бота от ОтцаБОтов
	Ratio               int      // количество сообщений между ответами бота
	Length              int      // длина сообщений генератора цепей
	EnableSemen         bool     // включить генерацию фраз
	EnablePhrases       bool     //включить фиксированные фразы
	MainCum             string   // ник владельца
	DefaultPhrases      []string // список фраз
	DefaultDataFileName string   // текстовый файл из которого берутся базовые данные
}
type Chat struct {
	ChatName       string
	CanTalkSemen   bool
	CanTalkPhrases bool
	Ratio          int //количество сообщений между ответами бота
	Counter        int //счетчик сообщений в чате
	SemenLength    int
	FixedPhrases   []string
	Cums           []string
}

type Bot struct {
	BotApi *tgbotapi.BotAPI
	Timer  time.Time
	Pause  time.Duration
	Cfg    Config

	Chats map[int64]*Chat

	Swatter map[int64]*swatter.DataStorage
}

func NewBot() *Bot {
	bot := Bot{}
	bot.LoadConfig()
	log.Println(bot.Cfg)
	if bot.Cfg.BotId == "" {
		panic("error creating new bot")
	}
	bapi, err := tgbotapi.NewBotAPI(bot.Cfg.BotId)
	if err != nil {
		log.Println("id: ", bot.Cfg.BotId)
		log.Fatal("starting tg bot error: ", err)
		return nil
	}
	bot.BotApi = bapi
	bot.Swatter = map[int64]*swatter.DataStorage{}
	bot.Chats = map[int64]*Chat{}
	bot.LoadDump()
	bot.Pause = 15 * time.Second
	bot.Timer = time.Now().UTC().Add(bot.Pause)
	return &bot
}

func (bot *Bot) Update(done <-chan bool) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.BotApi.GetUpdatesChan(u)
	go func() {
		for update := range updates {
			select {
			case <-done:
				return
			default:
			}
			bot.ShowUpdateInfo(update)
			if update.Message != nil && update.Message.Text != "" {
				bot.CheckChatSettings(update)
				if update.Message.IsCommand() {
					bot.Commands(update)
				} else {
					bot.ProcessMessage(update)
				}
			}
		}
	}()
}

func (bot *Bot) CheckChatSettings(update tgbotapi.Update) {
	_, ok := bot.Chats[update.FromChat().ID]
	// если впервые в чате, зададим дефолтный конфиг
	if !ok {
		bot.Chats[update.FromChat().ID] = &Chat{
			ChatName:       update.FromChat().Title,
			CanTalkSemen:   bot.Cfg.EnableSemen,
			CanTalkPhrases: bot.Cfg.EnablePhrases,
			Ratio:          bot.Cfg.Ratio,
			Counter:        0,
			SemenLength:    bot.Cfg.Length,
			FixedPhrases:   bot.Cfg.DefaultPhrases,
			Cums:           []string{bot.Cfg.MainCum},
		}
		var err error
		bot.Swatter[update.FromChat().ID], err = swatter.NewFromTextFile(bot.Cfg.DefaultDataFileName)
		if err != nil {
			log.Fatal("Eror creating new swatter ", err)
		}
		bot.SaveDump()
	}
	bot.Chats[update.FromChat().ID].ChatName = update.FromChat().Title
}

func (bot *Bot) ProcessMessage(update tgbotapi.Update) {
	chat := bot.Chats[update.FromChat().ID]
	chat.Counter++
	isTimeToTalk := chat.Ratio == 0 || (chat.Counter > chat.Ratio && bot.Tick()) //|| bot.IsCum(update.Message)
	if update.FromChat().IsPrivate() {
		msg := bot.GenerateMessage(update.Message)
		if msg == nil {
			return
		}
		bot.SendMessage(msg)
		return
	}
	if isTimeToTalk && chat.CanTalkPhrases {
		bot.SendFixedPhrase(update.Message)
		chat.Counter = 0
	} else if chat.CanTalkSemen {
		isReply := update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.UserName == bot.BotApi.Self.UserName
		if isTimeToTalk || isReply || bot.BotApi.IsMessageToMe(*update.Message) {
			chat.Counter = 0
			msg := bot.GenerateMessage(update.Message)
			if msg == nil {
				return
			}
			if !isTimeToTalk {
				switch concrete := msg.(type) {
				case tgbotapi.MessageConfig:
					concrete.ReplyToMessageID = update.Message.MessageID
					bot.SendMessage(concrete)
				case tgbotapi.StickerConfig:
					concrete.ReplyToMessageID = update.Message.MessageID
					bot.SendMessage(concrete)
				default:
					log.Println("ошибка")
				}
			} else {
				bot.SendMessage(msg)
			}
		}
	}
	bot.Learning(update.Message)
}

func (bot *Bot) IsCum(message *tgbotapi.Message) bool {
	mmb, err := bot.BotApi.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID:             message.Chat.ID,
			SuperGroupUsername: "",
			UserID:             message.From.ID,
		},
	})
	if err == nil {
		chat := bot.Chats[message.Chat.ID]
		for _, v := range chat.Cums {
			if v == mmb.User.UserName {
				return true
			}
		}
	}
	return false
}

func (bot *Bot) ShowUpdateInfo(update tgbotapi.Update) {
	if update.Message != nil && update.FromChat() != nil {
		if update.Message.NewChatMembers != nil {
			log.Println("new chat member for " + update.FromChat().Title + " " + update.Message.NewChatMembers[0].FirstName + " " + update.Message.NewChatMembers[0].UserName)
			//log.Println(update.Message.NewChatMembers)
		}
		if update.Message.LeftChatMember != nil {
			log.Println("user left from " + update.FromChat().Title + " " + update.Message.NewChatMembers[0].FirstName + " " + update.Message.NewChatMembers[0].UserName)
			log.Println(update.Message.LeftChatMember)
		}
	}
	if update.MyChatMember != nil && update.FromChat() != nil {
		log.Println("my chat member update " + update.FromChat().Title)
		log.Println(update.MyChatMember)
	}
}

func (bot *Bot) GenerateMessage(message *tgbotapi.Message) tgbotapi.Chattable {
	msg := bot.Swatter[message.Chat.ID].GenerateText(message.Text, bot.Chats[message.Chat.ID].SemenLength)
	//threadId := 0
	//if message.Chat.IsForum && message.MessageThreadID != 0 {
	//	threadId = message.MessageThreadID
	//}
	return tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: message.Chat.ID,
			//MessageThreadID:  threadId,
			ReplyToMessageID: 0,
		},
		Text:                  msg,
		DisableWebPagePreview: false,
	}
	//else {
	//    return tgbotapi.NewSticker(message.Chat.ID, tgbotapi.FileID(nefren))
	//}
}

func (bot *Bot) Learning(message *tgbotapi.Message) {
	bot.Swatter[message.Chat.ID].ParseText(message.Text)
}

func (bot *Bot) SendMessage(message tgbotapi.Chattable) {
	switch concrete := message.(type) {
	case tgbotapi.MessageConfig:
		if concrete.Text == "" {
			return
		}
	}
	_, err := bot.BotApi.Send(message)
	if err != nil {
		log.Println("Error sending message: ", err)
	}
}

func (bot *Bot) Reply(text string, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyToMessageID = message.MessageID
	bot.SendMessage(msg)
}

func (bot *Bot) ReplyNefren(message *tgbotapi.Message) {
	msg := tgbotapi.NewSticker(message.Chat.ID, tgbotapi.FileID(nefren))
	msg.ReplyToMessageID = message.MessageID
	bot.SendMessage(msg)
}

func (bot *Bot) SendFixedPhrase(message *tgbotapi.Message) {
	chat := bot.Chats[message.Chat.ID]
	if len(chat.FixedPhrases) != 0 {
		//threadId := 0
		//if message.Chat.IsForum && message.MessageThreadID != 0 {
		//	threadId = message.MessageThreadID
		//}
		bot.SendMessage(tgbotapi.MessageConfig{
			BaseChat: tgbotapi.BaseChat{
				ChatID: message.Chat.ID,
				//MessageThreadID:  threadId,
				ReplyToMessageID: 0,
			},
			Text:                  chat.FixedPhrases[rand.Intn(len(chat.FixedPhrases)-1)],
			DisableWebPagePreview: false,
		})
	}
}

func (bot *Bot) Tick() bool {
	isReady := time.Now().UTC().After(bot.Timer)
	if isReady {
		bot.Timer = time.Now().UTC().Add(bot.Pause)
	}
	return isReady
}

func (bot *Bot) LoadConfig() {
	log.Println("reading config...")
	content, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	err = json.Unmarshal(content, &bot.Cfg)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}
	log.Println("reading config...done")
}

func (bot *Bot) SaveConfig() {
	log.Println("saving config...")
	cfgJson, _ := json.Marshal(bot.Cfg)
	err := os.WriteFile("config.json", cfgJson, 0644)
	if err != nil {
		log.Fatal("Error during saving config: ", err)
	}
	log.Println("saving config...done")
}

func (bot *Bot) AddCum(chat *Chat, str string) int {
	if str != "" {
		chat.Cums = append(chat.Cums, str)
		bot.SaveDump()
		return len(chat.Cums) - 1
	}
	return -1
}
func (bot *Bot) AddFixedPhrase(chat *Chat, str string) int {
	if str != "" {
		chat.FixedPhrases = append(chat.FixedPhrases, str)
		bot.SaveDump()
		return len(chat.FixedPhrases) - 1
	}
	return -1
}

func (bot *Bot) RemoveFixedPhrase(chat *Chat, id int) int {
	if id > -1 && id < len(chat.FixedPhrases) {
		chat.FixedPhrases = append(chat.FixedPhrases[:id], chat.FixedPhrases[id+1:]...)
		bot.SaveDump()
		return len(chat.FixedPhrases)
	}
	return -1
}

func (bot *Bot) SaveDump() {
	cfgJson, _ := json.Marshal(bot.Chats)
	err := os.WriteFile("chats.json", cfgJson, 0644)
	if err != nil {
		log.Fatal("Error during saving chats: ", err)
	}
	for key, chat := range bot.Chats {
		bot.Swatter[key].SaveDump(chat.ChatName + ".blob")
	}
}

func (bot *Bot) SaveDumpFromTextFile(filepath string) {
	cfgJson, _ := json.Marshal(bot.Chats)
	err := os.WriteFile("chats.json", cfgJson, 0644)
	if err != nil {
		log.Fatal("Error during saving chats: ", err)
	}

	for key, chat := range bot.Chats {
		bot.Swatter[key].ReadFile(filepath)
		bot.Swatter[key].SaveDump(chat.ChatName + ".blob")
	}
}

func (bot *Bot) LoadDump() {
	log.Println("reading chats...")
	content, err := os.ReadFile("chats.json")
	if err != nil {
		cfgJson, _ := json.Marshal(bot.Chats)
		err := os.WriteFile("chats.json", cfgJson, 0644)
		if err != nil {
			log.Fatal("Error during saving chats: ", err)
		}
		log.Println("Error when opening file: ", err)
		return
	}
	err = json.Unmarshal(content, &bot.Chats)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	var needToSave bool
	for key, chat := range bot.Chats {
		bot.Swatter[key], err = swatter.NewFromDump(chat.ChatName + ".blob")
		if err != nil {
			bot.Swatter[key], err = swatter.NewFromTextFile(bot.Cfg.DefaultDataFileName)
			if err != nil {
				log.Fatal("Error creating new swatter ", err)
			}
			needToSave = true
		}
	}
	if needToSave {
		bot.SaveDump()
	}
	log.Println("reading chats...done")
}

func (bot *Bot) FixChats() {
	for id, c := range bot.Chats {
		chat, err := bot.BotApi.GetChat(tgbotapi.ChatInfoConfig{ChatConfig: tgbotapi.ChatConfig{ChatID: id, SuperGroupUsername: ""}})
		if err == nil {
			//log.Println(chat)
			if chat.IsPrivate() {
				log.Println("deleting " + c.ChatName)
				delete(bot.Chats, id)
			}
			log.Println("title " + chat.Title)
			c.ChatName = chat.Title
		} else {
			log.Println("deleting err ")
			delete(bot.Chats, id)
		}
	}
	bot.SaveDump()
}

func (bot *Bot) Commands(update tgbotapi.Update) {
	if bot.IsCum(update.Message) {
		chat := bot.Chats[update.FromChat().ID]
		switch update.Message.Command() {
		case commandAddCum:
			i := bot.AddCum(chat, update.Message.CommandArguments())
			bot.Reply("id:"+strconv.Itoa(i), update.Message)
			bot.SaveDump()
		case commandEnableSemen:
			chat.CanTalkSemen = update.Message.CommandArguments() == "да"
			bot.SaveDump()
		case commandEnablePhrases:
			chat.CanTalkPhrases = update.Message.CommandArguments() == "да"
			bot.SaveDump()
		case commandFixed:
			i := bot.AddFixedPhrase(chat, update.Message.CommandArguments())
			bot.Reply("id:"+strconv.Itoa(i), update.Message)
		case commandFixedRemove:
			id, err := strconv.Atoi(strings.TrimSpace(update.Message.CommandArguments()))
			if err != nil {
				bot.Reply(strconv.Itoa(len(chat.FixedPhrases)), update.Message)
			} else {
				i := bot.RemoveFixedPhrase(chat, id)
				bot.Reply("left:"+strconv.Itoa(i), update.Message)
				bot.SaveDump()
			}
		case commandRatio:
			ratio, err := strconv.Atoi(strings.TrimSpace(update.Message.CommandArguments()))
			if err != nil {
				bot.Reply(strconv.Itoa(chat.Ratio), update.Message)
			} else {
				chat.Ratio = ratio
				bot.SaveDump()
			}
		case commandLength:
			length, err := strconv.Atoi(strings.TrimSpace(update.Message.CommandArguments()))
			if err != nil {
				bot.Reply(strconv.Itoa(chat.SemenLength), update.Message)
			} else {
				chat.SemenLength = mathutil.Clamp(length, 1, maxLength)
				bot.SaveDump()
			}
		default:
			//bot.Reply("не понял" + update.Message.Command(), update.Message)
			log.Println(update.Message.Command())
		}
	}
}

func (bot *Bot) Dumper(done <-chan bool) {
	ticker := time.NewTicker(dumpTick)
	go func() {
		for {
			select {
			case <-done:
				bot.BotApi.StopReceivingUpdates()
				return
			case <-ticker.C:
				bot.SaveDump()
			}
		}
	}()
}

func (bot *Bot) Clean() {
	for key, _ := range bot.Chats {
		bot.Swatter[key].Clean()
	}
}
