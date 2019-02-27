package utils

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

var p *bluemonday.Policy

func init() {
	p = bluemonday.UGCPolicy()
	p.AllowAttrs("id").Matching(bluemonday.Integer).OnElements("span")
	p.AllowAttrs("class").Matching(regexp.MustCompile(`^reply$`)).OnElements("span")
}

func HTMLAndReplies(body string) (string, []int64) {
	unsafeWithReplies, replies := handleReplies(body)
	options := blackfriday.WithExtensions(blackfriday.CommonExtensions | blackfriday.HardLineBreak)
	unsafe := blackfriday.Run([]byte(unsafeWithReplies), options)
	html := p.SanitizeBytes(unsafe)
	return string(html), replies
}

func handleReplies(html string) (string, []int64) {
	// make unique map, empty struct occupies no additinal space
	uniqueReplies := make(map[int64]struct{})
	r := regexp.MustCompile(`(>>|<<)[0-9]+`)
	h := r.ReplaceAllStringFunc(html, func(s string) string {
		id, _ := strconv.ParseInt(s[2:], 10, 64)
		uniqueReplies[id] = struct{}{}
		// Added \n after each reply to enable blockqouting in next line
		return fmt.Sprintf(`<span id="%d" class="reply">%s</span>`+"\n", id, s)
	})
	var replies []int64
	for id := range uniqueReplies {
		replies = append(replies, id)
	}
	return h, replies
}
