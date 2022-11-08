package cybercum

import (
    "github.com/dyvdev/cybercum/tgbot"
    "time"
    "log"
    "strconv"
)

func ReadBot(cfgFile string) {
    log.Println("starting...")
    bot := tgbot.NewBot(cfgFile)
    log.Println("reading...")
    bot.Semen.ReadFile("reading")
    log.Println("saving...")
    bot.Semen.SaveDump(bot.Cfg.SaveFile)
    log.Println("done...")
}

func RunBot(cfgFile string) {
    log.Println("starting...")
    bot := tgbot.NewBot(cfgFile)
    c := make(chan int)
    go saver(c, bot)
    bot.Update()
    i := <-c
    log.Println("exit ", i)
}

func saver(c chan int, bot *tgbot.Bot) {
    for {
        time.Sleep(60 * 60 * time.Second)
        log.Println("saving.. [" + strconv.Itoa(len(bot.Semen)) + "]")
        bot.Semen.SaveDump(bot.Cfg.SaveFile)
    }
    c <- 1
}
