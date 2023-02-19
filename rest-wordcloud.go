package main

import (
	"database/sql"
	"flag"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/psykhi/wordclouds"
	"gopkg.in/yaml.v2"

	wordclass "rest-wordcloud/class"
)

var path = flag.String("input", "input.yaml", "path to flat YAML like {\"word\":42,...}")
var config = flag.String("config", "config.yaml", "path to config file")
var output = flag.String("output", "./images/cloud.png", "path to output image")
var cpuprofile = flag.String("cpuprofile", "profile", "write cpu profile to file")
var sqlpath = flag.String("sqlpath", "./db/wordCount.db", "path to sqlite database")

var listenPort = os.Getenv("LISTEN_PORT")

var DefaultColors = []color.RGBA{
	{0x1b, 0x1b, 0x1b, 0xff},
	{0x48, 0x48, 0x4B, 0xff},
	{0x59, 0x3a, 0xee, 0xff},
	{0x65, 0xCD, 0xFA, 0xff},
	{0x70, 0xD6, 0xBF, 0xff},
}

type Conf struct {
	FontMaxSize     int    `yaml:"font_max_size"`
	FontMinSize     int    `yaml:"font_min_size"`
	RandomPlacement bool   `yaml:"random_placement"`
	FontFile        string `yaml:"font_file"`
	Colors          []color.RGBA
	BackgroundColor color.RGBA `yaml:"background_color"`
	Width           int
	Height          int
	Mask            MaskConf
	SizeFunction    *string `yaml:"size_function"`
	Debug           bool
}

type MaskConf struct {
	File  string
	Color color.RGBA
}

var DefaultConf = Conf{
	FontMaxSize:     200, //300 //700
	FontMinSize:     3,   //4 // 10
	RandomPlacement: false,
	FontFile:        "./fonts/roboto/Roboto-Regular.ttf",
	Colors:          DefaultColors,
	BackgroundColor: color.RGBA{255, 255, 255, 255},
	Width:           1024, //2048
	Height:          1024, //2048
	Mask: MaskConf{"", color.RGBA{
		R: 0,
		G: 0,
		B: 0,
		A: 0,
	}},
	Debug: false,
}

func postWord(context *gin.Context) {
	var newWord wordclass.Words

	inputWord := context.Param("inputWord")

	if addWordToDb(inputWord, *sqlpath) {
		newWord.Word = inputWord
		publishWordcloud(context)
		context.IndentedJSON(http.StatusCreated, newWord)
	} else {
		context.IndentedJSON(http.StatusBadRequest, "Error adding word")
	}
}

func addWordToDb(wordToAdd string, filename string) bool {
	// TODO: Check if the file exists
	db, err := sql.Open("sqlite3", filename)

	if err != nil {
		panic(err)
	}

	var validChars = regexp.MustCompile(`\W+`)
	var cleanWord string

	fmt.Println(validChars.MatchString(wordToAdd))

	if validChars.MatchString(wordToAdd) == false {
		fmt.Println("Ordet inneholder ingen dritt")
		cleanWord = wordToAdd
	} else {
		fmt.Println("Ordet inneholder bare DRITT!!!")
		return false
	}

	var dbWord string
	var dbCount int

	dbQuery := fmt.Sprintf("SELECT * FROM wordcount WHERE word = '%s' LIMIT 1", cleanWord)
	rows, err := db.Query(dbQuery)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Ordet er: %s\n", cleanWord)

	if rows.Next() {
		//
		fmt.Println("Den her finnes fra før.")
		rows.Scan(&dbWord, &dbCount)
		rows.Close()
		db.Close()
		return updateWordCount(dbWord, dbCount, filename)
	} else {
		//
		fmt.Println("Denne var ny gitt.")
		rows.Close()
		db.Close()
		return insertNewWord(cleanWord, filename)
	}
}

func updateWordCount(word string, count int, filename string) bool {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		panic(err)
	}

	dbQuery := fmt.Sprintf("UPDATE wordcount SET count = %v WHERE word = '%s'", count+1, word)
	fmt.Printf("Ka kjøre vi: %s\n", dbQuery)
	defer db.Close()
	_, err = db.Exec(dbQuery)
	if err != nil {
		panic(err)
	}
	return true
}

func insertNewWord(word string, filename string) bool {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		panic(err)
	}

	dbQuery := fmt.Sprintf("INSERT INTO wordcount (word, count) VALUES ('%s', %v)", word, 1)
	defer db.Close()
	_, err = db.Exec(dbQuery)
	if err != nil {
		panic(err)
	}
	return true
}

func startUpMsg() {
	fmt.Println("")
	fmt.Println("__        __            _  ____ _                 _ ")
	fmt.Println("\\ \\      / /__  _ __ __| |/ ___| | ___  _   _  __| |")
	fmt.Println(" \\ \\ /\\ / / _ \\| '__/ _` | |   | |/ _ \\| | | |/ _` |")
	fmt.Println("  \\ V  V / (_) | | | (_| | |___| | (_) | |_| | (_| |")
	fmt.Println("   \\_/\\_/ \\___/|_|  \\__,_|\\____|_|\\___/ \\__,_|\\__,_|")
	fmt.Println("")
}

func main() {
	startUpMsg()
	var serverPort = 9090
	if port, err := strconv.Atoi(listenPort); err == nil {
		if port > 0 && port <= 65353 {
			serverPort = port
		}
	}
	router := gin.Default()
	router.SetTrustedProxies([]string{"192.168.1.0/24"})
	router.GET("/add/:inputWord", postWord)
	router.StaticFile("/cloud", "./images/cloud.png")
	router.StaticFile("/", "./index.html")
	router.Run("0.0.0.0:" + fmt.Sprintf("%v", serverPort))
}

func publishWordcloud(context *gin.Context) {
	var file string
	if *sqlpath != "" {
		file = "./db/wordCount.db"
	} else {
		file = *sqlpath
	}

	var dbWords []wordclass.Words = getWordsFromDb(file)

	if len(dbWords) <= 0 {
		log.Fatal("Database empty, no wordcloud without words.....")
		panic("Database empty, no wordcloud without words.....")
	}

	flag.Parse()
	/*
		if *cpuprofile != "" {
			f, err := os.Create(*cpuprofile)
			if err != nil {
				log.Fatal(err)
			}
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}
	*/
	inputWords := make(map[string]int, 0)

	for _, value := range dbWords {
		inputWords[value.Word] = value.Count
	}

	// Load config
	conf := DefaultConf
	configContent, err := os.ReadFile(*config)
	if err == nil {
		err = yaml.Unmarshal(configContent, &conf)
		if err != nil {
			fmt.Printf("Failed to decode config, using defaults instead: %s\n", err)
		}
	} else {
		fmt.Println("No config file. Using defaults")
	}
	os.Chdir(filepath.Dir(*config))

	if conf.Debug {
		confYaml, _ := yaml.Marshal(conf)
		fmt.Printf("Configuration: %s\n", confYaml)
	}

	var boxes []*wordclouds.Box
	if conf.Mask.File != "" {
		boxes = wordclouds.Mask(
			conf.Mask.File,
			conf.Width,
			conf.Height,
			conf.Mask.Color)
	}

	colors := make([]color.Color, 0)
	for _, c := range conf.Colors {
		colors = append(colors, c)
	}

	start := time.Now()
	oarr := []wordclouds.Option{wordclouds.FontFile(conf.FontFile),
		wordclouds.FontMaxSize(conf.FontMaxSize),
		wordclouds.FontMinSize(conf.FontMinSize),
		wordclouds.Colors(colors),
		wordclouds.MaskBoxes(boxes),
		wordclouds.Height(conf.Height),
		wordclouds.Width(conf.Width),
		wordclouds.RandomPlacement(conf.RandomPlacement),
		wordclouds.BackgroundColor(conf.BackgroundColor)}
	if conf.SizeFunction != nil {
		oarr = append(oarr, wordclouds.WordSizeFunction(*conf.SizeFunction))
	}
	if conf.Debug {
		oarr = append(oarr, wordclouds.Debug())
	}
	w := wordclouds.NewWordcloud(inputWords,
		oarr...,
	)

	img := w.Draw()
	outputFile, err := os.Create(*output)
	if err != nil {
		panic(err)
	}

	// Encode takes a writer interface and an image interface
	// We pass it the File and the RGBA
	png.Encode(outputFile, img)

	// Don't forget to close files
	outputFile.Close()

	fmt.Printf("Done in %v", time.Since(start))

}

func getWordsFromDb(filename string) []wordclass.Words {
	// TODO: Check if the file exists
	db, err := sql.Open("sqlite3", filename)

	if err != nil {
		panic(err)
	}

	dbQuery := "SELECT * FROM wordcount"

	rows, err := db.Query(dbQuery)

	if err != nil {
		panic(err)
	}

	defer rows.Close()
	var words = []wordclass.Words{}
	for rows.Next() {
		var word string
		var count int

		err = rows.Scan(&word, &count)
		if err != nil {
			panic(err)
		}

		var todoRow = []wordclass.Words{
			{
				Word:  word,
				Count: count,
			},
		}
		words = append(words, todoRow...)
	}

	return words
}
