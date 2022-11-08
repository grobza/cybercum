package tgbot

import (
    "log"
    "github.com/dyvdev/cybercum/semen"
    "math/rand"
    "regexp"
    //"encoding/json"
    "strings"
    "time"
    "strconv"
    "io/ioutil"
    "encoding/json"
    mathutil "modernc.org/mathutil"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Config struct {
    FileName string
    BotId string
    BotName string
    SaveFile string
    Ratio int
    Length int
}

type Bot struct {
    BotApi *tgbotapi.BotAPI
    Semen semen.Semen
    Timer time.Time
    Pause time.Duration
    Cfg Config
}

const (
    command_ratio = "ratio"
    command_length = "length"
    max_ratio = 100000
    max_length = 10

    nefren = "CAACAgIAAx0CTK3KYQACAQNjDKmYViPp5K-PWxuUKUDpwg0vQQAC9hEAAqx6iEqOhkQYAe2vbSkE"
)

func NewBot(cfgFile string) *Bot {
    bot := Bot{}
    bot.LoadConfig(cfgFile)
    log.Println(bot.Cfg)
    if bot.Cfg.BotId == "" || bot.Cfg.BotName == "" || bot.Cfg.FileName == ""{
        panic ("error creating new bot")
    }
    bapi, err := tgbotapi.NewBotAPI(bot.Cfg.BotId)
    if err != nil {
        log.Println("id: ", bot.Cfg.BotId)
        log.Fatal("starting tg bot error: ", err)
        return nil
    }
    bot.BotApi = bapi
    bot.Semen = *semen.NewFromDump(bot.Cfg.SaveFile)
    bot.Pause = 15 * time.Second
    bot.Timer = time.Now().UTC().Add(bot.Pause)
    return &bot
}

func (bot *Bot) Update() {
    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60
    updates := bot.BotApi.GetUpdatesChan(u)
    for update := range updates {
        if update.Message != nil {
            if update.Message.NewChatMembers != nil {
                log.Println(update.Message.NewChatMembers)
            }
            if update.Message.LeftChatMember != nil {
                log.Println(update.Message.LeftChatMember)
            }
        }
        if update.Message != nil && update.Message.Text != "" {
            if update.Message.IsCommand() {
                switch update.Message.Command() {
                case command_ratio:
                    ratio, err := strconv.Atoi(strings.TrimSpace(update.Message.CommandArguments()))
                    if err != nil {
                        bot.Reply(strconv.Itoa(bot.Cfg.Ratio) + "/" + strconv.Itoa(max_ratio), update.Message)
                    } else {
                        bot.Cfg.Ratio = mathutil.Clamp(ratio, 0, max_ratio)
                        bot.SaveConfig()
                    }
                case command_length:
                    length, err := strconv.Atoi(strings.TrimSpace(update.Message.CommandArguments()))
                    if err != nil {
                        bot.Reply(strconv.Itoa(bot.Cfg.Length), update.Message)
                    } else {
                        bot.Cfg.Length = mathutil.Clamp(length, 1, max_length)
                        bot.SaveConfig()
                    }
                default:
                    //bot.Reply("не понял" + update.Message.Command(), update.Message)
                    log.Println(update.Message.Command())
                }
            } else {
                isReply := update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.UserName == bot.Cfg.BotName
                isMention := strings.Contains(update.Message.Text, "@" + bot.Cfg.BotName)
                isTimeToTalk := (rand.Intn(max_ratio) < bot.Cfg.Ratio) && bot.Tick()

                if  isMention || isReply || isTimeToTalk {
                    msg := bot.GenerateMessage(update.Message)
                    if !isTimeToTalk {
                        switch concrete := msg.(type) {
                        case tgbotapi.MessageConfig:
                            concrete.ReplyToMessageID = update.Message.MessageID
                            bot.BotApi.Send(concrete)
                        case tgbotapi.StickerConfig:
                            concrete.ReplyToMessageID = update.Message.MessageID
                            bot.BotApi.Send(concrete)
                        default:
                            log.Println("ошибка")
                        }
                    } else {
                        bot.BotApi.Send(msg)
                    }
                }
            }
        }
    }
}

func (bot *Bot) GenerateMessage(message *tgbotapi.Message) tgbotapi.Chattable {
    reg, err := regexp.Compile(`\p{Cyrillic}+`)
    if err != nil {
        log.Fatal(err)
    }
    words := reg.FindAllString(message.Text, -1)
    if len(words) != 0 {
        text := bot.Semen.Talk(words[rand.Intn(len(words))], bot.Cfg.Length)
        bot.Semen.Learning(words)
        return tgbotapi.NewMessage(message.Chat.ID, text)
    } else {
        return tgbotapi.NewSticker(message.Chat.ID, tgbotapi.FileID(nefren))
    }
}

func (bot Bot) Reply(text string, message *tgbotapi.Message) {
    msg := tgbotapi.NewMessage(message.Chat.ID, text)
    msg.ReplyToMessageID = message.MessageID
    bot.BotApi.Send(msg)
}

func (bot Bot) ReplyNefren(message *tgbotapi.Message) {
    msg:= tgbotapi.NewSticker(message.Chat.ID, tgbotapi.FileID(nefren))
    msg.ReplyToMessageID = message.MessageID
    bot.BotApi.Send(msg)
}

func (bot *Bot) Tick() bool {
    isReady := time.Now().UTC().After(bot.Timer)
    if isReady {
        bot.Timer = time.Now().UTC().Add(bot.Pause)
    }
    return isReady
}

func (bot *Bot) LoadConfig(cfgFile string) {
    log.Println("reading config...")
    content, err := ioutil.ReadFile(cfgFile)
    if err != nil {
        log.Fatal("Error when opening file: ", err)
    }
    err = json.Unmarshal(content, &bot.Cfg)
    if err != nil {
        log.Fatal("Error during Unmarshal(): ", err)
    }
    log.Println("reading config...done")
}

func (bot Bot) SaveConfig() {
    log.Println("saving config...")
    cfgJson, _ := json.Marshal(bot.Cfg)
    err := ioutil.WriteFile(bot.Cfg.FileName, cfgJson, 0644)
    if err != nil {
        log.Fatal("Error during saving config: ", err)
    }
    log.Println("saving config...done")
}
