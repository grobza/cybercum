package main

import (
	"flag"
	cum "github.com/dyvdev/cybercum"
	"github.com/dyvdev/cybercum/swatter"
	"github.com/dyvdev/cybercum/utils"
	"log"
	"math/rand"
	"os"
	"time"
)

var chatHistoryPath *string
var configPath *string

func init() {
	chatHistoryPath = flag.String("chat history dump path", "d", "")
	configPath = flag.String("config path", "c", "")
}

func main() {
	rand.Seed(time.Now().UnixNano())
	cum.RunBot(*configPath)
	//test()
	//testChat()
}

func testChat() {
	sw := &swatter.DataStorage{}
	data := utils.GetTgData("tghistory.json")
	file, err := os.Create("tghistory.txt")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()
	for _, str := range data {
		sw.ParseText(str)
		file.WriteString(str)
	}
	log.Print(sw.GenerateText("", 15))
}

func test() {
	sw := &swatter.DataStorage{}
	sw.ReadFile("mh.txt")

	log.Print(sw.GenerateText("кум", 5))
	log.Print(sw.GenerateText("рома", 10))
	log.Print(sw.GenerateText("да", 15))
	log.Print(sw.GenerateText("нет", 25))
}
