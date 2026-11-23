package doubanhar

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"dinosaur/internal/history"
	"golang.org/x/net/html"
)

var subject = regexp.MustCompile(`^/subject/([1-9][0-9]*)/$`)
var count = regexp.MustCompile(`\(([0-9]+)\)`)
var year = regexp.MustCompile(`^([0-9]{4})`)
var rating = regexp.MustCompile(`^rating([1-5])-t$`)

func attr(n *html.Node, key string) string {
	if n == nil {
		return ""
	}
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}
func hasClass(n *html.Node, name string) bool {
	for _, c := range strings.Fields(attr(n, "class")) {
		if c == name {
			return true
		}
	}
	return false
}
func walk(n *html.Node, visit func(*html.Node)) {
	if n == nil {
		return
	}
	visit(n)
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		walk(child, visit)
	}
}
func find(n *html.Node, predicate func(*html.Node) bool) *html.Node {
	if n == nil {
		return nil
	}
	if predicate(n) {
		return n
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if found := find(child, predicate); found != nil {
			return found
		}
	}
	return nil
}
func text(n *html.Node) string {
	if n == nil {
		return ""
	}
	var parts []string
	walk(n, func(node *html.Node) {
		if node.Type == html.TextNode {
			parts = append(parts, node.Data)
		}
	})
	return strings.Join(strings.Fields(strings.Join(parts, " ")), " ")
}
func class(n *html.Node, name string) *html.Node {
	return find(n, func(node *html.Node) bool { return hasClass(node, name) })
}

func parsePage(body string) ([]Record, int, error) {
	root, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	total := 0
	heading := find(root, func(n *html.Node) bool { return n.Data == "h1" })
	if match := count.FindStringSubmatch(text(heading)); match != nil {
		total, _ = strconv.Atoi(match[1])
	}
	items := []Record{}
	var parseErr error
	walk(root, func(n *html.Node) {
		if parseErr != nil || !hasClass(n, "item") || !hasClass(n, "comment-item") {
			return
		}
		record, err := parseItem(n)
		if err != nil {
			parseErr = err
			return
		}
		items = append(items, record)
	})
	return items, total, parseErr
}

func parseItem(n *html.Node) (Record, error) {
	e := history.Entry{Source: "douban"}
	link := find(n, func(node *html.Node) bool {
		u, err := url.Parse(attr(node, "href"))
		return err == nil && u.Scheme == "https" && u.Hostname() == "movie.douban.com" && u.User == nil && subject.MatchString(u.Path)
	})
	if link == nil {
		return Record{}, fmt.Errorf("watched item has no Douban subject")
	}
	u, _ := url.Parse(attr(link, "href"))
	e.SourceID = subject.FindStringSubmatch(u.Path)[1]
	title := class(n, "title")
	em := find(title, func(node *html.Node) bool { return node.Data == "em" })
	titleText := text(em)
	if titleText == "" {
		titleText = text(title)
	}
	parts := strings.SplitN(titleText, " / ", 2)
	e.Title = parts[0]
	if len(parts) == 2 {
		e.OriginalTitle = parts[1]
	}
	e.WatchedOn = text(class(n, "date"))
	e.Review = text(class(n, "comment"))
	if match := year.FindStringSubmatch(text(class(n, "intro"))); match != nil {
		e.Year, _ = strconv.Atoi(match[1])
	}
	walk(n, func(node *html.Node) {
		for _, c := range strings.Fields(attr(node, "class")) {
			if match := rating.FindStringSubmatch(c); match != nil {
				e.Rating, _ = strconv.Atoi(match[1])
			}
		}
	})
	if err := e.Validate(); err != nil {
		return Record{}, fmt.Errorf("invalid Douban record %s: %w", e.SourceID, err)
	}
	image := find(n, func(node *html.Node) bool { return node.Data == "img" && posterURL(attr(node, "src")) })
	poster := ""
	if image != nil {
		poster = attr(image, "src")
	}
	return Record{Entry: e, PosterURL: poster}, nil
}

func fmtScanStart(raw string, dst *int) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("invalid collection page offset")
	}
	*dst = value
	return value, nil
}
