// Package web serves the HTML interface.
package web

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/amila-t-kumarasekara/web-analyzer/internal/analyzer"
	"github.com/amila-t-kumarasekara/web-analyzer/internal/fetcher"
	"github.com/amila-t-kumarasekara/web-analyzer/internal/linkcheck"
	"github.com/amila-t-kumarasekara/web-analyzer/internal/validate"
	"github.com/amila-t-kumarasekara/web-analyzer/pkg/consts"
)

//go:embed templates/index.html
var templateFS embed.FS

// Small interfaces, defined where they are used.
type pageFetcher interface {
	Fetch(ctx context.Context, u *url.URL) (*fetcher.Page, error)
}

type linkChecker interface {
	Check(ctx context.Context, urls []string) linkcheck.Result
}

// Handler serves the form and analysis results.
type Handler struct {
	fetcher pageFetcher
	checker linkChecker
	tmpl    *template.Template
	log     *slog.Logger
}

// pageView is what the template renders.
type pageView struct {
	URL           string
	Result        *analyzer.Result
	HeadingRows   []headingRow
	Broken        []linkcheck.Outcome
	LinkCheckNote string
	Err           *viewError
}

type viewError struct {
	StatusCode  int
	StatusText  string
	Description string
}

// NewHandler builds the router with all routes and middleware.
func NewHandler(f pageFetcher, c linkChecker) http.Handler {
	tmpl, err := template.ParseFS(templateFS, "templates/index.html")
	if err != nil {
		panic("web: parse templates: " + err.Error())
	}

	h := &Handler{
		fetcher: f,
		checker: c,
		tmpl:    tmpl,
		log:     slog.Default(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", h.index)
	mux.HandleFunc("POST /analyze", h.analyze)
	mux.HandleFunc("GET /healthz", h.healthz)

	return chain(mux)
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	h.render(w, pageView{})
}

func (h *Handler) analyze(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.render(w, pageView{
			Err: &viewError{
				StatusCode:  http.StatusBadRequest,
				StatusText:  http.StatusText(http.StatusBadRequest),
				Description: "Invalid form submission.",
			},
		})
		return
	}

	raw := r.FormValue("url")
	u, err := validate.URL(raw)
	if err != nil {
		h.render(w, pageView{
			URL: raw,
			Err: &viewError{
				StatusCode:  http.StatusBadRequest,
				StatusText:  http.StatusText(http.StatusBadRequest),
				Description: validate.Description(raw, err),
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), consts.AnalyzeTimeout)
	defer cancel()

	page, err := h.fetcher.Fetch(ctx, u)
	if err != nil {
		h.render(w, h.pageFromFetchErr(raw, err))
		return
	}

	result, err := analyzer.Analyze(bytes.NewReader(page.Body), page.FinalURL)
	if err != nil {
		h.log.Error("analyze failed", slog.Any("err", err))
		h.render(w, pageView{
			URL: raw,
			Err: &viewError{
				StatusCode:  http.StatusInternalServerError,
				StatusText:  http.StatusText(http.StatusInternalServerError),
				Description: "Could not parse the page HTML.",
			},
		})
		return
	}

	linkCtx, linkCancel := context.WithTimeout(r.Context(), consts.LinkCheckTimeout)
	defer linkCancel()
	checked := h.checker.Check(linkCtx, result.CheckURLs)
	broken := make([]linkcheck.Outcome, 0)
	for _, o := range checked.Outcomes {
		if !o.OK {
			broken = append(broken, o)
		}
	}

	var linkNote string
	if checked.Incomplete {
		linkNote = fmt.Sprintf(
			"Link check timed out: checked %d of %d links. Counts for inaccessible links may be low.",
			len(checked.Outcomes), checked.Total,
		)
	}

	h.render(w, pageView{
		URL:           raw,
		Result:        &result,
		HeadingRows:   headingRows(result.Headings),
		Broken:        broken,
		LinkCheckNote: linkNote,
	})
}

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (h *Handler) render(w http.ResponseWriter, view pageView) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.Execute(w, view); err != nil {
		h.log.Error("render template", slog.Any("err", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (h *Handler) pageFromFetchErr(raw string, err error) pageView {
	var fe *fetcher.Error
	if errors.As(err, &fe) {
		status := fe.StatusCode
		if status == 0 {
			status = http.StatusBadGateway
		}
		return pageView{
			URL: raw,
			Err: &viewError{
				StatusCode:  status,
				StatusText:  http.StatusText(status),
				Description: fe.Reason,
			},
		}
	}
	return pageView{
		URL: raw,
		Err: &viewError{
			StatusCode:  http.StatusInternalServerError,
			StatusText:  http.StatusText(http.StatusInternalServerError),
			Description: "Could not fetch the page.",
		},
	}
}

type headingRow struct {
	Level string
	Count int
}

func headingRows(counts [6]int) []headingRow {
	rows := make([]headingRow, 0, 6)
	for i, n := range counts {
		if n == 0 {
			continue
		}
		rows = append(rows, headingRow{
			Level: fmt.Sprintf("h%d", i+1),
			Count: n,
		})
	}
	return rows
}
