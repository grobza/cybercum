package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	cum "github.com/dyvdev/cybercum"
	"github.com/dyvdev/cybercum/swatter"
	"github.com/dyvdev/cybercum/utils"
)

func main() {
	var histPath *string
	histPath = flag.String("rhist", "", "read chat dump, overwrite blob and exit")
	flag.Parse()
	if len(*histPath) != 0 {
		ChatHistoryGen(*histPath)
		return
	}
	rand.Seed(time.Now().UnixNano())
	cum.RunBot()
}

func ChatHistoryGen(historyFile string) {
	sw := &swatter.DataStorage{}
	data := utils.GetTgData(historyFile)
	baseName := filepath.Base(historyFile)
	cleanName := strings.Split(baseName, ".")
	if len(cleanName) < 2 {
		return
	}

	textDumpName := cleanName[0] + ".txt"
	file, err := os.Create(textDumpName)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()
	for _, str := range data {
		newStr := sw.ParseText(str)
		if strings.Contains(str, "порядке") {
			str = str + " "
		}
		if newStr == "" {
			continue
		}

		file.WriteString(newStr + " ")
	}
}

func test() {
	sw := &swatter.DataStorage{}
	sw.ReadFile("mh.txt")

	log.Print(sw.GenerateText("кум", 5))
	log.Print(sw.GenerateText("рома", 10))
	log.Print(sw.GenerateText("да", 15))
	log.Print(sw.GenerateText("нет", 25))
}
