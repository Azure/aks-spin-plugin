package logger

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	// verbose describes whether the logger should output information that is typically useful for debugging
	verbose = false
)

// SetVerbose sets the verbosity of the logger. Verbose being true results in all logs being printed to user which
// typically is useful for debugging.
func SetVerbose(v bool) {
	verbose = v
}

func new() *slog.Logger {
	return slog.New(&handler{
		out: os.Stdout,
		mu:  &sync.Mutex{},
	})
}

// handler is a custom slog handler that prints output in a cli-friendly format and adheres to the
// verbose global
type handler struct {
	attrs  []slog.Attr // data from WithAttrs
	groups []string    // data from WithGroup
	mu     *sync.Mutex
	out    io.Writer
}

func (h *handler) Enabled(ctx context.Context, l slog.Level) bool {
	if verbose {
		return true
	}

	return l.Level() >= slog.LevelInfo
}

// Handle prints the logging message in a digestable format for cli users.
func (h *handler) Handle(ctx context.Context, r slog.Record) error {
	output := []interface{}{r.Message}

	if r.Level > slog.LevelWarn {
		output = append([]interface{}{r.Level.String()}, output...)
	}

	r.Attrs(func(a slog.Attr) bool {
		output = append(output, a)
		return true
	})

	for _, attr := range h.attrs {
		output = append(output, attr)
	}

	if groups := h.groups; len(groups) != 0 {
		output = append(output, slog.String("group", strings.Join(groups, "-")))
	}

	// also only print timestamps when verbose
	if verbose && !r.Time.IsZero() {
		output = append(output, slog.String(slog.TimeKey, r.Time.Format("2006-01-02T15:04:05.999Z")))
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	log.New(h.out, "", 0).Println(output...)

	return nil
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	h2 := *h
	h2.attrs = make([]slog.Attr, len(h.attrs))
	copy(h2.attrs, h.attrs)
	h2.attrs = append(h2.attrs, attrs...)

	return &h2
}

func (h *handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	h2 := *h
	h2.groups = make([]string, len(h.groups)+1)
	copy(h2.groups, h.groups)
	h2.groups[len(h2.groups)-1] = name

	return &h2
}
