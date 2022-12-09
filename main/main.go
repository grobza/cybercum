package main

import (
	cum "github.com/dyvdev/cybercum"
	"github.com/dyvdev/cybercum/swatter"
	"github.com/dyvdev/cybercum/utils"
	"log"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	//cum.ReadBot("./config.json")
	cum.RunBot("./config.json")
	//test()
	//testChat()
}

func testChat() {
	sw := &swatter.DataStorage{}
	data := utils.GetTgData("don.json")
	for _, str := range data {
		sw.ParseText(str)
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
