package ingest

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

// OPMLFeed is a feed entry extracted from an OPML file.
type OPMLFeed struct {
	URL   string
	Label string
}

// LoadOPML reads RSS feed URLs from an OPML subscription file.
func LoadOPML(path string) ([]OPMLFeed, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseOPML(data)
}

// ParseOPML extracts outline entries with xmlUrl from OPML XML.
func ParseOPML(data []byte) ([]OPMLFeed, error) {
	var doc opmlDocument
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse opml: %w", err)
	}
	var feeds []OPMLFeed
	collectOutlines(doc.Body.Outlines, &feeds)
	if len(feeds) == 0 {
		return nil, fmt.Errorf("opml: no rss outlines with xmlUrl found")
	}
	return feeds, nil
}

type opmlDocument struct {
	Body opmlBody `xml:"body"`
}

type opmlBody struct {
	Outlines []opmlOutline `xml:"outline"`
}

type opmlOutline struct {
	Type     string        `xml:"type,attr"`
	Title    string        `xml:"title,attr"`
	XMLURL   string        `xml:"xmlUrl,attr"`
	Outlines []opmlOutline `xml:"outline"`
}

func collectOutlines(nodes []opmlOutline, feeds *[]OPMLFeed) {
	for _, n := range nodes {
		url := strings.TrimSpace(n.XMLURL)
		if url != "" {
			label := strings.TrimSpace(n.Title)
			*feeds = append(*feeds, OPMLFeed{URL: url, Label: label})
		}
		if len(n.Outlines) > 0 {
			collectOutlines(n.Outlines, feeds)
		}
	}
}
