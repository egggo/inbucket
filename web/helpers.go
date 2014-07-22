package web

import (
	"crypto/md5"
	"fmt"
	"github.com/egggo/inbucket/log"
	"html"
	"html/template"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var TemplateFuncs = template.FuncMap{
	"friendlyTime": friendlyTime,
	"reverse":      reverse,
	"textToHtml":   textToHtml,
}

// From http://daringfireball.net/2010/07/improved_regex_for_matching_urls
var urlRE = regexp.MustCompile("(?i)\\b((?:[a-z][\\w-]+:(?:/{1,3}|[a-z0-9%])|www\\d{0,3}[.]|[a-z0-9.\\-]+[.][a-z]{2,4}/)(?:[^\\s()<>]+|\\(([^\\s()<>]+|(\\([^\\s()<>]+\\)))*\\))+(?:\\(([^\\s()<>]+|(\\([^\\s()<>]+\\)))*\\)|[^\\s`!()\\[\\]{};:'\".,<>?«»“”‘’]))")

// Friendly date & time rendering
func friendlyTime(t time.Time) template.HTML {
	ty, tm, td := t.Date()
	ny, nm, nd := time.Now().Date()
	if (ty == ny) && (tm == nm) && (td == nd) {
		return template.HTML(t.Format("03:04:05 PM"))
	}
	return template.HTML(t.Format("Mon Jan 2, 2006"))
}

// Reversable routing function (shared with templates)
func reverse(name string, things ...interface{}) string {
	// Convert the things to strings
	strs := make([]string, len(things))
	for i, th := range things {
		strs[i] = fmt.Sprint(th)
	}
	// Grab the route
	u, err := Router.Get(name).URL(strs...)
	if err != nil {
		log.LogError("Failed to reverse route: %v", err)
		return "/ROUTE-ERROR"
	}
	return u.Path
}

// textToHtml takes plain text, escapes it and tries to pretty it up for
// HTML display
func textToHtml(text string) template.HTML {
	text = html.EscapeString(text)
	text = urlRE.ReplaceAllStringFunc(text, wrapUrl)
	replacer := strings.NewReplacer("\r\n", "<br/>\n", "\r", "<br/>\n", "\n", "<br/>\n")
	return template.HTML(replacer.Replace(text))
}

// wrapUrl wraps a <a href> tag around the provided URL
func wrapUrl(url string) string {
	unescaped := strings.Replace(url, "&amp;", "&", -1)
	return fmt.Sprintf("<a href=\"%s\" target=\"_blank\">%s</a>", unescaped, url)
}

// 计算 生成md5crypt 盐值
func salt() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	magic := strconv.FormatInt(r.Int63(), 10)
	md5magic := fmt.Sprintf("%X", md5.Sum([]byte(magic)))
	return md5magic[0:8]
}
