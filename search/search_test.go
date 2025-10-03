package search

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func Test_regex(t *testing.T) {
	input := `<!-- This is a comment --> Some text <!-- Another comment --> More text`
	re := regexp.MustCompile(`<!--(.*?)-->`)
	// matches := re.FindAllString(input, -1)
	matches := re.FindAllStringSubmatch(input, 2)
	var match string
	for _, matchGroup := range matches {
		if len(matchGroup) > 1 {
			match = strings.Trim(matchGroup[1], " ")
			fmt.Println(match)
		}
	}
	fmt.Println(match)
}

// func Test_parseDate(t *testing.T) {
// 	type args struct {
// 		date time.Time
// 		path string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    string
// 		wantErr bool
// 	}{
// 		{
// 			name: "first_essay.md",
// 			args: args{
// 				date: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
// 				path: "/docs/essays/first_essay.md",
// 			},
// 			want: "2024-12-02",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			os.Chdir("/home/thinker/Documents/Programs/Golang/go-blog")
// 			got, err := parseDate(tt.args.date, tt.args.path)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("parseDate() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if got != tt.want {
// 				t.Errorf("parseDate() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_NewIndex(t *testing.T) {
// 	os.Chdir("..")
// 	NewIndex(200)
// }

// func Test_Search(t *testing.T) {
// 	os.Chdir("..")
// 	index := NewIndex(200)
// 	// index := LoadIndex()
// 	docs, err := index.Search([]string{"massa", "nunct"})
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	for _, doc := range docs {
// 		fmt.Println(doc.Name, doc.Score)
// 	}
// }

// func Test_get_ngrams(t *testing.T) {
// 	words := []string{"the", "quick", "brown", "fox", "jumps", "over", "the", "lazy", "dog"}
// 	ng := ngrams(words, 3)
// 	for _, n := range ng {
// 		fmt.Println(n)
// 	}
// }

// func Test_NewDoc(t *testing.T) {
// 	os.Chdir("..")
// 	dirs, err := os.ReadDir("./docs")
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	docOpts := DocOpts{
// 		LoadContent: true,
// 		LenPreview:  200,
// 	}

// 	var docList []Document
// 	for _, dir := range dirs {
// 		files, err := os.ReadDir("./docs/" + dir.Name())
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		for _, file := range files {
// 			info, err := file.Info()
// 			if err != nil {
// 				t.Error(err)
// 			}
// 			loc := Location{
// 				Info: info,
// 				Dir:  dir,
// 			}
// 			doc, err := loc.NewDoc(docOpts)
// 			// doc, err := NewDoc(info, dir, false)
// 			if err != nil {
// 				t.Error(err)
// 			}
// 			docList = append(docList, doc)
// 		}
// 	}

// 	for _, doc := range docList {
// 		fmt.Println(doc.Name, doc.Path, doc.Date, doc.Preview)
// 	}
// }
