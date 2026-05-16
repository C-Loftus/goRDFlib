package nt

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	rdflibgo "github.com/tggo/goRDFlib"
	"github.com/tggo/goRDFlib/internal/ntsyntax"
)

// TripleHandler is the callback used by ParseStream. Returning a non-nil error
// aborts the parse and is propagated to the caller of ParseStream.
type TripleHandler func(s rdflibgo.Subject, p rdflibgo.URIRef, o rdflibgo.Term) error

// Parse parses N-Triples format RDF into the given graph.
func Parse(g *rdflibgo.Graph, r io.Reader, opts ...Option) error {
	return parseLines(r, opts, func(s rdflibgo.Subject, p rdflibgo.URIRef, o rdflibgo.Term) error {
		g.Add(s, p, o)
		return nil
	})
}

// ParseStream parses N-Triples format RDF and dispatches each parsed triple to
// the handler without populating any graph. Use this for streaming large inputs
// where holding the full graph in memory is not feasible. Returning an error
// from the handler aborts the parse.
func ParseStream(r io.Reader, h TripleHandler, opts ...Option) error {
	if h == nil {
		return fmt.Errorf("nt.ParseStream: handler must not be nil")
	}
	return parseLines(r, opts, h)
}

// parseLines is the shared scanner loop used by Parse and ParseStream.
func parseLines(r io.Reader, opts []Option, h TripleHandler) error {
	var cfg config
	for _, o := range opts {
		o(&cfg)
	}
	scanner := bufio.NewScanner(r)
	if cfg.maxLineLen > 0 {
		scanner.Buffer(make([]byte, 0, 64*1024), cfg.maxLineLen)
	}
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		if err := parseNTLine(line, lineNum, h); err != nil {
			if cfg.errorHandler == nil {
				return err
			}
			fixedLine, retry := cfg.errorHandler(lineNum, line, err)
			if retry {
				if err2 := parseNTLine(fixedLine, lineNum, h); err2 != nil {
					return fmt.Errorf("line %d: retry failed: %w", lineNum, err2)
				}
			}
		}
	}
	return scanner.Err()
}

func parseNTLine(line string, lineNum int, h TripleHandler) error {
	p := &ntsyntax.LineParser{Line: line, Pos: 0, LineNum: lineNum}

	subj, err := p.ReadSubject()
	if err != nil {
		return err
	}
	p.SkipSpaces()

	pred, err := p.ReadPredicate()
	if err != nil {
		return err
	}
	p.SkipSpaces()

	obj, err := p.ReadObject()
	if err != nil {
		return err
	}
	p.SkipSpaces()

	if !p.Expect('.') {
		return fmt.Errorf("line %d: expected '.'", lineNum)
	}

	return h(subj, pred, obj)
}
