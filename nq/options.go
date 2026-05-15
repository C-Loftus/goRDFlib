package nq

// Option configures N-Quads parsing or serialization.
type Option func(*config)

// ErrorHandler is called when a line fails to parse.
// It receives the 1-based line number, the raw line text, and the parse error.
// If retry is true, fixedLine is parsed instead (exactly one retry attempt).
// If retry is false, the line is skipped and parsing continues.
// To preserve the default fail-fast behavior, do not set an error handler.
type ErrorHandler func(lineNum int, line string, err error) (fixedLine string, retry bool)

type config struct {
	base         string
	quadHandler  QuadHandler
	errorHandler ErrorHandler
	maxLineLen   int
}

// WithBase sets the base IRI for resolving relative IRIs.
func WithBase(base string) Option {
	return func(c *config) { c.base = base }
}

// WithQuadHandler sets a callback that receives the graph context for each parsed quad.
// The graph term is nil for triples without an explicit graph context.
func WithQuadHandler(h QuadHandler) Option {
	return func(c *config) { c.quadHandler = h }
}

// WithErrorHandler sets a callback invoked when a line fails to parse.
// See ErrorHandler for semantics.
func WithErrorHandler(h ErrorHandler) Option {
	return func(c *config) { c.errorHandler = h }
}

// WithMaxLineLength raises the maximum byte length of a single N-Quads line.
// The default (bufio.MaxScanTokenSize, 64KB) is too small for inputs that pack
// large literals — e.g. WKT polygons — onto one line. Pass a value larger than
// the longest expected line in bytes.
func WithMaxLineLength(n int) Option {
	return func(c *config) { c.maxLineLen = n }
}
