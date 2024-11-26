package slog2

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	_, b, _, _ = runtime.Caller(0)
	root       = filepath.Join(filepath.Dir(b), "../..") + "/"
)

type PlainSlogHandler struct {
	opts Options
	mu   *sync.Mutex
	goas []groupOrAttrs
	out  io.Writer
}

type Options struct {
	AddSource bool
	Level     slog.Leveler
}

var levelText = map[slog.Level]string{
	slog.LevelDebug: "[debug]",
	slog.LevelInfo:  "[info] ",
	slog.LevelWarn:  "[warn] ",
	slog.LevelError: "[error]",
}

const unknownText = "[????]  "

// groupOrAttrs holds either a group name or a list of slog.Attrs.
type groupOrAttrs struct {
	group string      // group name if non-empty
	attrs []slog.Attr // attrs if non-empty
}

func New(out io.Writer, opts *Options) *PlainSlogHandler {
	h := &PlainSlogHandler{
		out: out,
		mu:  &sync.Mutex{},
	}
	if opts != nil {
		h.opts = *opts
	}
	if h.opts.Level == nil {
		h.opts.Level = slog.LevelInfo
	}
	return h
}

func (h *PlainSlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *PlainSlogHandler) withGroupOrAttrs(goa groupOrAttrs) *PlainSlogHandler {
	h2 := *h
	l := len(h.goas)
	h2.goas = make([]groupOrAttrs, l+1)
	copy(h2.goas, h.goas)
	h2.goas[l] = goa
	return &h2
}

func (h *PlainSlogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return h.withGroupOrAttrs(groupOrAttrs{group: name})
}

func (h *PlainSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	return h.withGroupOrAttrs(groupOrAttrs{attrs: attrs})
}

const (
	timeUTC   = "15:04:05.000000"
	timeLocal = "15:04:05.000000-0700"
)

func init() {
	tzEnv := os.Getenv("TZ")
	if tzEnv != "" {
		tz, _ := time.LoadLocation(tzEnv)
		if tz != nil {
			time.Local = tz
		}
	}
}

func (h *PlainSlogHandler) Handle(ctx context.Context, r slog.Record) error {
	buf := make([]byte, 0, 1024)

	if !r.Time.IsZero() {
		buf = fmt.Append(buf, r.Time.Format(time.DateOnly), " ")
		if r.Time.Location() == nil || r.Time.Location() == time.UTC {
			buf = fmt.Append(buf, r.Time.Format(timeUTC), " ")
		} else {
			buf = fmt.Append(buf, r.Time.Format(timeLocal), " ")
		}
	}

	if lvl, ok := levelText[r.Level]; ok {
		buf = fmt.Append(buf, lvl)
	} else {
		buf = fmt.Append(buf, unknownText)
	}

	buf = fmt.Append(buf, " ", r.Message)

	if h.opts.AddSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		file := strings.TrimPrefix(f.File, root)
		buf = fmt.Append(buf, "  {", file, ":", f.Line, "}  ")
	}

	indentLevel := 0
	// Handle state from WithGroup and WithAttrs.
	goas := h.goas
	if r.NumAttrs() == 0 {
		// If the record has no Attrs, remove groups at the end of the list; they are empty.
		for len(goas) > 0 && goas[len(goas)-1].group != "" {
			goas = goas[:len(goas)-1]
		}
	}

	separate := true
	for _, goa := range goas {
		if goa.group != "" {
			buf = fmt.Append(buf, "  ", goa.group, "={ ")
			indentLevel++
			separate = false
		} else {
			for _, a := range goa.attrs {
				buf = h.appendAttr(buf, a, separate)
				separate = true
			}
		}
	}

	r.Attrs(func(a slog.Attr) bool {
		buf = h.appendAttr(buf, a, separate)
		separate = true
		return true
	})
	for i := 0; i < indentLevel; i++ {
		buf = append(buf, ' ', '}') // Close group.
	}
	buf = append(buf, '\n')
	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.out.Write(buf)
	return err
}

func (h *PlainSlogHandler) appendAttr(buf []byte, a slog.Attr, separate bool) []byte {
	// Resolve the Attr's value before doing anything else.
	a.Value = a.Value.Resolve()
	// Ignore empty Attrs.
	if a.Equal(slog.Attr{}) {
		return buf
	}
	if separate {
		buf = append(buf, ' ', ' ')
	}

	switch a.Value.Kind() {
	case slog.KindString:
		// Quote string values, to make them easy to parse.
		buf = fmt.Appendf(buf, "%s=%q", a.Key, a.Value.String())
	case slog.KindTime:
		// Write times in a standard way, without the monotonic time.
		buf = fmt.Append(buf, a.Key, "=", a.Value.Time().Format(time.RFC3339Nano))
	case slog.KindGroup:

		attrs := a.Value.Group()
		// Ignore empty groups.
		if len(attrs) == 0 {
			return buf
		}
		// If the key is non-empty, write it out and indent the rest of the attrs.
		// Otherwise, inline the attrs.
		if a.Key != "" {
			buf = fmt.Append(buf, a.Key, "=")
		}
		buf = append(buf, '{', ' ')
		for i, ga := range attrs {
			buf = h.appendAttr(buf, ga, i > 0)
		}
		buf = append(buf, ' ', '}')
	default:
		buf = fmt.Append(buf, a.Key, "=", a.Value)
	}
	return buf
}
