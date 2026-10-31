// Package nfo converts a bounded subset of Kodi movie/tvshow NFO into domain data.
package nfo

import (
	"bytes"
	"encoding/xml"
	"io"
	"sort"
	"strings"

	"dinosaur/internal/agent"
	"dinosaur/internal/metadata"
)

const MaxSize = 128 * 1024

type uniqueID struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}
type extensions struct {
	Language         string   `xml:"language,omitempty"`
	OriginalLanguage string   `xml:"original_language,omitempty"`
	SourceURL        string   `xml:"source_url,omitempty"`
	Aliases          []string `xml:"alias,omitempty"`
}
type document struct {
	XMLName       xml.Name
	Title         string     `xml:"title"`
	OriginalTitle string     `xml:"originaltitle,omitempty"`
	Overview      string     `xml:"plot,omitempty"`
	Year          int        `xml:"year,omitempty"`
	Premiered     string     `xml:"premiered,omitempty"`
	IDs           []uniqueID `xml:"uniqueid"`
	Extension     extensions `xml:"dinosaur"`
}

func Decode(data []byte, kind string) (agent.Input, error) {
	if len(data) > MaxSize {
		return agent.Input{}, agent.ErrInput
	}
	root := "movie"
	if kind == "tv" {
		root = "tvshow"
	} else if kind != "movie" {
		return agent.Input{}, agent.ErrInput
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	depth, roots, tokens := 0, 0, 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return agent.Input{}, agent.ErrInput
		}
		tokens++
		if tokens > 20000 {
			return agent.Input{}, agent.ErrInput
		}
		switch token := token.(type) {
		case xml.Directive:
			return agent.Input{}, agent.ErrInput
		case xml.ProcInst:
			if token.Target != "xml" {
				return agent.Input{}, agent.ErrInput
			}
		case xml.StartElement:
			if depth == 0 {
				roots++
				if token.Name.Local != root || token.Name.Space != "" {
					return agent.Input{}, agent.ErrInput
				}
			}
			depth++
			if depth > 32 || roots > 1 {
				return agent.Input{}, agent.ErrInput
			}
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && strings.TrimSpace(string(token)) != "" {
				return agent.Input{}, agent.ErrInput
			}
		}
	}
	if roots != 1 || depth != 0 {
		return agent.Input{}, agent.ErrInput
	}
	var d document
	if xml.Unmarshal(data, &d) != nil {
		return agent.Input{}, agent.ErrInput
	}
	ids := map[string]string{}
	for _, id := range d.IDs {
		provider := strings.ToLower(strings.TrimSpace(id.Type))
		if provider == "themoviedb" {
			provider = "tmdb"
		}
		if previous, exists := ids[provider]; exists && previous != strings.TrimSpace(id.Value) {
			return agent.Input{}, agent.ErrInput
		}
		ids[provider] = strings.TrimSpace(id.Value)
	}
	input := agent.Input{Title: d.Title, OriginalTitle: d.OriginalTitle, Overview: d.Overview, Year: d.Year, ReleaseDate: d.Premiered,
		Language: d.Extension.Language, OriginalLanguage: d.Extension.OriginalLanguage, SourceURL: d.Extension.SourceURL, Aliases: d.Extension.Aliases, IDs: ids}
	_, err := input.Movie("nfo")
	return input, err
}

func Encode(info metadata.Movie, kind string) ([]byte, error) {
	root := "movie"
	if kind == "tv" {
		root = "tvshow"
	} else if kind != "movie" {
		return nil, agent.ErrInput
	}
	input := agent.FromMovie(info)
	d := document{XMLName: xml.Name{Local: root}, Title: input.Title, OriginalTitle: input.OriginalTitle, Overview: input.Overview, Year: input.Year, Premiered: input.ReleaseDate,
		Extension: extensions{Language: input.Language, OriginalLanguage: input.OriginalLanguage, SourceURL: input.SourceURL, Aliases: input.Aliases}}
	providers := []string{}
	for provider := range input.IDs {
		providers = append(providers, provider)
	}
	sort.Strings(providers)
	for _, provider := range providers {
		d.IDs = append(d.IDs, uniqueID{Type: provider, Value: input.IDs[provider]})
	}
	data, err := xml.MarshalIndent(d, "", "  ")
	return append([]byte(xml.Header), data...), err
}
