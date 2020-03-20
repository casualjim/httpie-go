package output

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/go-xmlfmt/xmlfmt"
	"github.com/logrusorgru/aurora"
)

type PrettyPrinter struct {
	writer        io.Writer
	plain         Printer
	aurora        aurora.Aurora
	headerPalette *HeaderPalette
	indentWidth   int
}

type PrettyPrinterConfig struct {
	Writer      io.Writer
	EnableColor bool
}

type HeaderPalette struct {
	Method              aurora.Color
	URL                 aurora.Color
	Proto               aurora.Color
	SuccessfulStatus    aurora.Color
	NonSuccessfulStatus aurora.Color
	FieldName           aurora.Color
	FieldValue          aurora.Color
	FieldSeparator      aurora.Color
}

var defaultHeaderPalette = HeaderPalette{
	Method:              aurora.WhiteFg | aurora.BoldFm,
	URL:                 aurora.GreenFg | aurora.BoldFm,
	Proto:               aurora.BlueFg,
	SuccessfulStatus:    aurora.GreenFg | aurora.BoldFm,
	NonSuccessfulStatus: aurora.YellowFg | aurora.BoldFm,
	FieldName:           aurora.WhiteFg,
	FieldValue:          aurora.CyanFg,
	FieldSeparator:      aurora.WhiteFg,
}

func NewPrettyPrinter(config PrettyPrinterConfig) Printer {
	return &PrettyPrinter{
		writer:        config.Writer,
		plain:         NewPlainPrinter(config.Writer),
		aurora:        aurora.NewAurora(config.EnableColor),
		headerPalette: &defaultHeaderPalette,
		indentWidth:   4,
	}
}

func (p *PrettyPrinter) PrintStatusLine(resp *http.Response) error {
	var statusColor aurora.Color
	if resp.StatusCode/100 == 2 {
		statusColor = p.headerPalette.SuccessfulStatus
	} else {
		statusColor = p.headerPalette.NonSuccessfulStatus
	}

	fmt.Fprintf(p.writer, "%s %s\n",
		p.aurora.Colorize(resp.Proto, p.headerPalette.Proto),
		p.aurora.Colorize(resp.Status, statusColor))
	return nil
}

func (p *PrettyPrinter) PrintRequestLine(req *http.Request) error {
	fmt.Fprintf(p.writer, "%s %s %s\n",
		p.aurora.Colorize(req.Method, p.headerPalette.Method),
		p.aurora.Colorize(req.URL, p.headerPalette.URL),
		p.aurora.Colorize(req.Proto, p.headerPalette.Proto))
	return nil
}

func (p *PrettyPrinter) PrintHeader(header http.Header) error {
	var names []string
	for name := range header {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		values := header[name]
		for _, value := range values {
			fmt.Fprintf(p.writer, "%s%s %s\n",
				p.aurora.Colorize(name, p.headerPalette.FieldName),
				p.aurora.Colorize(":", p.headerPalette.FieldSeparator),
				p.aurora.Colorize(value, p.headerPalette.FieldValue))
		}
	}

	fmt.Fprintln(p.writer)
	return nil
}

func cleanContentType(contentType string) string {
	contentType = strings.TrimSpace(contentType)

	semicolon := strings.Index(contentType, ";")
	if semicolon != -1 {
		contentType = contentType[:semicolon]
	}
	return contentType
}

func isJSON(contentType string) bool {
	return cleanContentType(contentType) == "application/json"
}

func isXML(contentType string) bool {
	return cleanContentType(contentType) == "application/xml"
}

func (p *PrettyPrinter) PrintBody(body io.Reader, contentType string) error {
	l := lexers.MatchMimeType(contentType)
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	// Determine formatter.
	f := formatters.TTY8
	if os.Getenv("TERM") == "xterm-256color" {
		f = formatters.TTY256
	}
	if os.Getenv("COLORTERM") == "truecolor" {
		f = formatters.TTY16m
	}

	// Determine style.
	s := styles.Dracula
	if os.Getenv("HT_THEME") != "" {
		ss := styles.Get(os.Getenv("HT_THEME"))
		if ss != nil {
			s = ss
		}
	}

	var source string
	if isJSON(contentType) {
		var data interface{}
		if err := json.NewDecoder(body).Decode(&data); err != nil {
			return fmt.Errorf("decoding json: %w", err)
		}
		bb, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("re-encoding json: %w", err)
		}
		source = string(bb)
	} else {
		bb, err := ioutil.ReadAll(body)
		if err != nil {
			return fmt.Errorf("reading body: %w", err)
		}
		source = string(bb)
		if isXML(contentType) {
			source = xmlfmt.FormatXML(source, "", "  ")
		}
	}

	it, err := l.Tokenise(nil, source)
	if err != nil {
		return err
	}
	return f.Format(p.writer, s, it)
}
