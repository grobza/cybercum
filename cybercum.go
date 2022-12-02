package cybercum

import (
	"github.com/dyvdev/cybercum/tgbot"
	"log"
	"time"
)

func ReadBot(cfgFile string) {
	log.Println("starting...")
	bot := tgbot.NewBot()
	log.Println("reading...")
	bot.Semen.ReadFile("reading")
	log.Println("saving...")
	bot.SaveDump()
	log.Println("done...")
}

func RunBot(cfgFile string) {
	log.Println("starting...")
	bot := tgbot.NewBot()
	c := make(chan int)
	go saver(c, bot)
	bot.Update()
	i := <-c
	log.Println("exit ", i)
}

func saver(c chan int, bot *tgbot.Bot) {
	for {
		time.Sleep(60 * 60 * time.Second)
		bot.SaveDump()
	}
	c <- 1
}
