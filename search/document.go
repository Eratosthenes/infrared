package search

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"example.com/jlogger"
	"example.com/lexer"
)

type Location struct {
	log *jlogger.JLog

	Info os.FileInfo
	Dir  os.DirEntry
	Err  error
}

type DocOpts struct {
	WithContent bool
	LenPreview  int
}

type Document struct {
	length int

	Name     string  `json:"name"`
	Date     string  `json:"date"`
	Preview  string  `json:"preview"`   // first 100 characters, using ellipsis if truncated
	Path     string  `json:"path"`      // markdown file path
	HtmlPath string  `json:"html_path"` // html file path
	Score    float64 // score for a given search
	Content  string  // full content, lowercase
}

func NewLocation(jlog *jlogger.JLog, dir os.DirEntry, info os.FileInfo) *Location {
	return &Location{
		Dir:  dir,
		Info: info,
		log:  jlog,
	}
}

func (l *Location) NewDoc(opts DocOpts) (Document, error) {
	d := Document{}
	actions := [](func(*Document, DocOpts)){
		l.setName,
		l.setPath,
		l.setHtmlPath,
		l.setDate,
		l.setContent,
		l.setPreview,
	}
	for _, action := range actions {
		action(&d, opts)
		if l.Err != nil {
			return Document{}, l.Err
		}
	}
	return d, nil
}

// populates the index with the documents in the docs directory
func GetDocs(lenPreview int) ([]Document, error) {
	var docs []Document
	dirs, err := os.ReadDir("./docs")
	if err != nil {
		return []Document{}, err
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		files, err := os.ReadDir("./docs/" + dir.Name())
		if err != nil {
			return []Document{}, err
		}

		for _, file := range files {
			info, err := file.Info()
			if err != nil {
				return []Document{}, err
			}

			loc := NewLocation(jlogger.New(), dir, info)
			doc, err := loc.NewDoc(DocOpts{
				WithContent: true,
				LenPreview:  lenPreview,
			})
			if err != nil {
				return []Document{}, err
			}
			docs = append(docs, doc)
		}
	}

	return docs, nil
}

func (l *Location) setName(d *Document, opts DocOpts) {
	if l.Err != nil {
		return
	}
	docName := strings.TrimSuffix(l.Info.Name(), ".md")
	docName = strings.ReplaceAll(docName, "_", " ")
	firstLetter := strings.ToUpper(string(docName[0]))
	docName = firstLetter + docName[1:]
	d.Name = docName
}

func (l *Location) setPath(d *Document, opts DocOpts) {
	if l.Err != nil {
		return
	}
	dirName, fileName := l.Dir.Name(), l.Info.Name()
	d.Path = fmt.Sprintf("/docs/%s/%s", dirName, fileName)
}

func (l *Location) setHtmlPath(d *Document, opts DocOpts) {
	if l.Err != nil {
		return
	}
	htmlName := strings.Split(l.Info.Name(), ".")[0] + ".html"
	htmlName = strings.ReplaceAll(htmlName, "'", "")
	dirName := l.Dir.Name()
	d.HtmlPath = fmt.Sprintf("/views/docs/%s/%s", dirName, htmlName)
}

func (l *Location) setDate(d *Document, opts DocOpts) {
	if l.Err != nil {
		return
	}
	date, err := parseDate(l.Info.ModTime(), d.Path)
	if err != nil {
		l.Err = err
		return
	}
	d.Date = date
}

func (l *Location) setContent(d *Document, opts DocOpts) {
	if l.Err != nil || !opts.WithContent {
		return
	}

	err := setContent(d)
	if err != nil {
		l.Err = err
		return
	}
}

func setContent(d *Document) error {
	contentBytes, err := os.ReadFile("." + d.Path)
	if err != nil {
		return err
	}
	parsedMd := lexer.ParseMarkdown(string(contentBytes))

	splitPunct := regexp.MustCompile(`[\s.,;:!?()\[\]]+`)
	punctuation := regexp.MustCompile(`[^a-zA-Z0-9\s-]`)

	var docWords []string
	for _, word := range parsedMd {
		word = splitPunct.ReplaceAllString(word, " ")
		word = punctuation.ReplaceAllString(word, "")
		docWords = append(docWords, strings.ToLower(word))
	}

	titleWords := strings.Fields(d.Name)
	for i := range titleWords {
		titleWords[i] = strings.ToLower(titleWords[i])
	}

	d.Content = strings.Join(append(titleWords, docWords...), " ")
	d.length = len(docWords)
	return nil
}

func (l *Location) setPreview(doc *Document, opts DocOpts) {
	if l.Err != nil {
		return
	}
	file, err := os.Open("." + doc.Path)
	if err != nil {
		return
	}
	defer file.Close()

	var words []string
	charCount := 0
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		word := scanner.Text()
		charCount += len(word)
		words = append(words, word)
		if charCount > opts.LenPreview {
			break
		}
	}
	contentStr := strings.Join(words, " ")
	if err := scanner.Err(); err != nil {
		return
	}
	parsedMd := lexer.ParseMarkdown(contentStr)
	parsedWords := truncate(parsedMd, opts.LenPreview)

	// skip the date if it is the first word
	previewWords := strings.Fields(parsedWords)
	if isDateFormat(previewWords[0]) {
		doc.Preview = strings.Join(previewWords[1:], " ")
		previewWords = previewWords[1:]
	}

	// skip the title if it's there
	match := true
	title := strings.Fields(doc.Name)
	lenTitle := len(strings.Fields(doc.Name))
	for i := 0; i < lenTitle; i++ {
		previewWord := strings.ToLower(previewWords[i])
		titleWord := strings.ToLower(title[i])
		if previewWord != titleWord {
			match = false
		}
	}
	if match {
		previewWords = previewWords[lenTitle:]
	}
	doc.Preview = strings.Join(previewWords, " ")
}

// isDateFormat checks if a string is in the format of a date
func isDateFormat(word string) bool {
	pattern := `^\d{4}-\d{2}-\d{2}$`
	matched, err := regexp.MatchString(pattern, word)
	if err != nil {
		return false
	}
	return matched
}

// truncate iterates through each word in the content until the lenPreview is reached, then appends an ellipsis
func truncate(content []string, lenPreview int) string {
	var tc string // truncated content
	for _, word := range content {
		if len(tc)+len(word) > lenPreview {
			tc = tc[:len(tc)-1]
			tc += " "
			break
		}
		tc += word + " "
	}
	return tc[:len(tc)-1] + "..."
}

func parseDate(date time.Time, path string) (string, error) {
	docBytes, err := os.ReadFile("." + path)
	if err != nil {
		return "", err
	}
	docString := string(docBytes)

	dateRx := regexp.MustCompile("<p style=\"color:gray\">(.*?)</p>")
	matchGroup := dateRx.FindStringSubmatch(docString)
	var dateString string
	if len(matchGroup) > 1 {
		dateString = strings.Trim(matchGroup[1], " ")
	}
	if dateString != "" {
		return dateString, nil
	}

	return strings.Split(date.String(), " ")[0], nil
}
