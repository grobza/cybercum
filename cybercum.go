package cybercum

import (
	"github.com/dyvdev/cybercum/tgbot"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func ReadBot(cfgFile string) {
	log.Println("starting...")
	bot := tgbot.NewBot()
	log.Println("reading...")
	//bot.Swatter.ReadFile("mh.txt")
	log.Println("saving...")
	bot.SaveDump()
	log.Println("done...")
}

func CleanBot() {
	bot := tgbot.NewBot()
	bot.Clean()
	bot.SaveDump()
}

func RunBot() {
	log.Println("starting...")
	done := make(chan bool)
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM)
	bot := tgbot.NewBot()
	bot.Update(done)
	bot.Dumper(done)
	<-sigc
	done <- true
}
