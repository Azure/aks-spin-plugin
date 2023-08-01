package logger

import (
	"context"
	"log"
	"os"

	"golang.org/x/exp/slog"
)

var def *slog.Logger

// TODO: make the WithAttrs and WithGroup actuall work
// TODO: set a way to set level and set it with a --verbose flag command passed to cli

// handler is a custom slog handler that prints output in a cli-friendly format.
type handler struct {
	handler slog.Handler
	l       *log.Logger
}

// Enabled implements slog.Handler.
func (h *handler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.handler.Enabled(ctx, l)
}

// Handle prints the logging message in a digestable format for cli users.
func (h *handler) Handle(ctx context.Context, r slog.Record) error {
	output := []interface{}{r.Message}
	r.Attrs(func(a slog.Attr) bool {
		output = append(output, a)
		return true
	})

	// we only print message and attributes.
	// timestamp and level aren't useful to cli users

	h.l.Println(output...)
	return nil
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &handler{
		handler: h.handler.WithAttrs(attrs),
		l:       h.l,
	}
}

func (h *handler) WithGroup(name string) slog.Handler {
	return &handler{
		handler: h.handler.WithGroup(name),
		l:       h.l,
	}
}

func getDef() *slog.Logger {
	if def != nil {
		return def
	}

	out := os.Stdout
	def = slog.New(&handler{
		handler: slog.NewTextHandler(out, nil),
		l:       log.New(out, "", 0),
	})
	return def
}
