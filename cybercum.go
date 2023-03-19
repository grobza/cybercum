package cybercum

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dyvdev/cybercum/tgbot"
)

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
