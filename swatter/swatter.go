package swatter

import (
	"bufio"
	"encoding/gob"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/dyvdev/cybercum/utils"
)

type Trigram [3]string
type DataStorage map[string]map[string]map[string]int

func ComposeTrigram(trigramMap DataStorage, msg string) Trigram {
	msg = strings.ToLower(regexp.MustCompile(`\.|,|;|!|\?`).ReplaceAllString(msg, ""))
	words := strings.Split(msg, " ")
	var trigram Trigram
	trigram[0] = utils.TrimWord(words[rand.Intn(len(words))])
	if trigram[0] == "" {
		r1 := rand.Intn(len(trigramMap))
		for word := range trigramMap {
			r1--
			if r1 <= 0 {
				trigram[0] = word
				break
			}
		}
	}
	if _, ok := trigramMap[trigram[0]]; !ok {
		return GetRandomTrigram(trigramMap)
	}
	r2 := rand.Intn(len(trigramMap[trigram[0]]))
	for word := range trigramMap[trigram[0]] {
		r2--
		if r2 <= 0 {
			trigram[1] = word
			break
		}
	}
	r3 := rand.Intn(len(trigramMap[trigram[0]][trigram[1]]))
	for word := range trigramMap[trigram[0]][trigram[1]] {
		r3--
		if r3 <= 0 {
			trigram[2] = word
			break
		}
	}
	return trigram
}

func GetRandomTrigram(data DataStorage) Trigram {
	var ret [3]string
	rnd := rand.Intn(len(data))
	for word := range data {
		rnd--
		if rnd <= 0 {
			ret[0] = word
			break
		}
	}
	rnd = rand.Intn(len(data[ret[0]]))
	for word := range data[ret[0]] {
		rnd--
		if rnd <= 0 {
			ret[1] = word
			break
		}
	}
	rnd = rand.Intn(len(data[ret[0]][ret[1]]))
	for word := range data[ret[0]][ret[1]] {
		rnd--
		if rnd <= 0 {
			ret[2] = word
			break
		}
	}
	return ret
}

func GetRandomWord(wordWeights map[string]int) string {
	dataWeight := 0
	for _, w := range wordWeights {
		dataWeight += w
	}
	weight := rand.Intn(dataWeight)
	for word, w := range wordWeights {
		dataWeight -= w
		if dataWeight <= weight {
			return word
		}
	}
	return ""
}

func (data DataStorage) AddTrigram(trigram Trigram) {
	first, ok := data[trigram[0]]
	if !ok {
		first = make(map[string]map[string]int)
		data[trigram[0]] = first
	}
	second, ok := first[trigram[1]]
	if !ok {
		second = make(map[string]int)
		first[trigram[1]] = second
	}
	second[trigram[2]]++
}

func (data DataStorage) GenerateText(msg string, length int) string {
	var last2Words [2]string
	var text []string
	if len(data) == 0 {
		return ""
	}
	if len(msg) > 0 {
		trigram := ComposeTrigram(data, msg)
		text = append(text, trigram[:]...)
		last2Words[0] = trigram[1]
		last2Words[1] = trigram[2]
	}
	for i := 0; i < length; i++ {
		if len(text) > 0 {
			possibleNextWords := data[last2Words[0]][last2Words[1]]
			if len(possibleNextWords) == 0 {
				break
			}
			nextWord := GetRandomWord(possibleNextWords)
			text = append(text, nextWord)
			last2Words[0] = last2Words[1]
			last2Words[1] = nextWord
		} else {
			trigram := GetRandomTrigram(data)
			text = append(text, trigram[:]...)
			last2Words[0] = trigram[1]
			last2Words[1] = trigram[2]
		}
	}
	for i := range text {
		text[i] = cleanWord(text[i])

	}
	return strings.Join(text, " ")
}

func cleanWord(word string) string {
	if len(word) < 2 {
		return ""
	}
	word = utils.TrimWord(word)
	if word[0] == '"' && word[len(word)-1] != '"' {
		word = strings.ReplaceAll(word, "\"", "")
	}
	if word[0] == '\'' && word[len(word)-1] != '\'' {
		word = strings.ReplaceAll(word, "'", "")
	}
	word = strings.Trim(word, "\"\\n\\")
	return word
}

func (data DataStorage) ParseText(text string) string {
	text = strings.ToLower(regexp.MustCompile(`\.|,|;|!|\?|\t`).ReplaceAllString(text, ""))
	text = strings.ToLower(regexp.MustCompile(`[^a-zA-Zа-яА-Я\s\d]`).ReplaceAllString(text, ""))
	text = strings.ToLower(regexp.MustCompile(`\n{2,}`).ReplaceAllString(text, "\n"))
	text = strings.ToLower(regexp.MustCompile(" {2,}").ReplaceAllString(text, " "))
	text = strings.ToLower(regexp.MustCompile(" {2,}").ReplaceAllString(text, " "))
	text = strings.TrimSpace(text)

	words := strings.Split(text, " ")
	var trimmedWords []string
	last := ""
	for _, word := range words {
		word = utils.TrimWord(word)
		if word != last {
			trimmedWords = append(trimmedWords, word)
		}
		last = word
	}
	for i := 0; i < len(trimmedWords)-2; i++ {
		if trimmedWords[i] != trimmedWords[i+2] {
			data.AddTrigram(Trigram{trimmedWords[i], trimmedWords[i+1], trimmedWords[i+2]})
		}
	}
	return text
}

func (data DataStorage) ReadFile(filename string) error {
	if filename == "" {
		return nil
	}
	f, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if err != nil {
		log.Println(err)
		return err
	}
	for scanner.Scan() {
		s := scanner.Text()
		data.ParseText(s)
	}
	return err
}
func (data DataStorage) SaveDump(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()
	e := gob.NewEncoder(file)
	if err = e.Encode(data); err != nil {
		log.Fatal(err)
		return
	}
}

func (data DataStorage) LoadDump(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := gob.NewDecoder(f)
	if err := dec.Decode(&data); err != nil {
		return err
	}
	return nil
}

func NewFromDump(filename string) (*DataStorage, error) {
	data := DataStorage{}
	return &data, data.LoadDump(filename)
}

func NewFromTextFile(filename string) (*DataStorage, error) {
	data := DataStorage{}
	return &data, data.ReadFile(filename)
}

func (data DataStorage) Clean() {
	counter1 := 0
	counter2 := 0
	for word1 := range data {
		for word2 := range data[word1] {
			var str []string
			for word3 := range data[word1][word2] {
				if word1 == word3 {
					str = append(str, word3)
					counter2++
				}
				counter1++
			}
			for _, s := range str {
				delete(data[word1][word2], s)
			}
			if len(data[word1][word2]) == 0 {
				delete(data[word1], word2)
				if len(data[word1]) == 0 {
					delete(data, word1)
				}
			}
		}
	}
	log.Println(strconv.Itoa(counter1) + " " + strconv.Itoa(counter2))
}
