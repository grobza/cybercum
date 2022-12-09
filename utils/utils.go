package utils

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"unicode"
)

func TrimWord(word string) string {
	mapping := map[rune]rune{
		'a': 'a',
		'b': 'б',
		'c': 'ц',
		'd': 'д',
		'e': 'е',
		'f': 'ф',
		'g': 'г',
		'h': 'х',
		'i': 'и',
		'j': 'ж',
		'k': 'к',
		'l': 'л',
		'm': 'м',
		'n': 'н',
		'o': 'о',
		'p': 'п',
		'q': 'к',
		'r': 'р',
		's': 'с',
		't': 'т',
		'u': 'ю',
		'v': 'в',
		'w': 'у',
		'x': 'х',
		'y': 'у',
		'z': 'з',
	}
	if strings.HasPrefix(word, "@") {
		word = strings.TrimPrefix(word, "@")
		var ret []rune
		for _, r := range word {
			switch {
			case unicode.Is(unicode.Latin, r):
				ret = append(ret, mapping[r])
			default:
				ret = append(ret, r)
			}
		}
		return string(ret)
	}
	return word
}

func GetTgData(filename string) []string {
	type chatData struct {
		Name     string
		Type     string
		Id       int
		Messages []struct {
			Id            int
			Type          string
			Date          string
			Date_unix     string
			Actor         string
			Actor_id      string
			Action        string
			Title         string
			Text          string
			Text_entities []struct {
				Type string
				Text string
			}
		}
	}
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	var data chatData
	err = json.Unmarshal(content, &data)
	if err != nil {
		log.Println("Error during Unmarshal(): ", err)
	}
	var ret []string
	for _, msg := range data.Messages {
		//log.Println(msg.Text)
		ret = append(ret, msg.Text)
	}
	return ret
}
