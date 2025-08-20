package templates

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Templates struct {
	mu      sync.RWMutex
	t       map[string]*template.Template
	embedFS embed.FS
	dev     bool
	path    string
}

func NewTemplates(fs embed.FS) *Templates {
	p := &Templates{
		mu:      sync.RWMutex{},
		t:       make(map[string]*template.Template),
		embedFS: fs,
		dev:     false,
		path:    "",
	}

	p.loadAllTemplates()

	return p
}

func NewDevTemplates(fs embed.FS, path string) *Templates {
	tmpl := NewTemplates(fs)

	tmpl.dev = true
	tmpl.path = path

	return tmpl
}

func (p *Templates) loadAllTemplates() {
	templates := make(map[string]*template.Template)
	var fragmentPaths []string

	err := fs.WalkDir(p.embedFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".html") {
			return nil
		}
		if !strings.Contains(path, "fragments/") {
			return nil
		}
		name := strings.TrimPrefix(path, "templates/")
		name = strings.TrimSuffix(name, ".html")
		tmpl, err := template.New(name).ParseFS(p.embedFS, path)
		if err != nil {
			slog.Error("failed setting up template fragment", "error", err)
			os.Exit(1)
		}
		templates[name] = tmpl
		fragmentPaths = append(fragmentPaths, path)
		slog.Debug("loaded fragment", "name", name)
		return nil
	})
	if err != nil {
		slog.Error("failed walking template dir for fragments", "error", err)
		os.Exit(1)
	}

	err = fs.WalkDir(p.embedFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".html") {
			return nil
		}
		if strings.Contains(path, "fragments/") {
			return nil
		}
		if strings.Contains(path, "layouts/") {
			return nil
		}
		name := strings.TrimPrefix(path, "templates/")
		name = strings.TrimSuffix(name, ".html")

		templatePaths := []string{}
		templatePaths = append(templatePaths, "templates/layouts/*.html")
		templatePaths = append(templatePaths, fragmentPaths...)
		templatePaths = append(templatePaths, path)

		tmpl, err := template.New(name).ParseFS(p.embedFS, templatePaths...)
		if err != nil {
			slog.Error("failed setting up template", "error", err)
			os.Exit(1)
		}
		templates[name] = tmpl
		slog.Debug("loaded template", "name", name)
		return nil
	})
	if err != nil {
		slog.Error("failed walking template dir", "error", err)
		os.Exit(1)
	}

	slog.Debug("templates loaded", "count", len(templates))
	p.mu.Lock()
	defer p.mu.Unlock()
	p.t = templates
}

func (p *Templates) loadOneFromDisk(path string) error {
	tmplPath := filepath.Join(p.path, path+".html")
	if _, err := os.Stat(tmplPath); os.IsNotExist(err) {
		return fmt.Errorf("template file not found on disk: %w", err)
	}

	var fragmentPaths []string
	err := fs.WalkDir(p.embedFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".html") {
			return nil
		}
		if !strings.Contains(path, "fragments/") {
			return nil
		}
		name := strings.TrimPrefix(path, "templates/")
		name = strings.TrimSuffix(name, ".html")
		fragmentPaths = append(fragmentPaths, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("error loading fragments from disk: %w", err)
	}

	tmpl := template.New(path)

	glob := filepath.Join(p.path, "layouts", "*.html")
	layouts, err := filepath.Glob(glob)
	if err != nil {
		return fmt.Errorf("error finding layout templates: %w", err)
	}

	files := append(layouts, fragmentPaths...)
	files = append(files, tmplPath)

	tmpl, err = tmpl.ParseFiles(files...)
	if err != nil {
		return fmt.Errorf("error parsing template files: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.t[path] = tmpl
	slog.Debug("template loaded from disk", "name", path)

	return nil
}

func (p *Templates) ExecutePlain(name string, w io.Writer, params any) error {
	return p.executeReload(name, w, "", params)
}

func (p *Templates) Execute(name string, w io.Writer, base string, params any) error {
	return p.executeReload(name, w, base, params)
}

func (p *Templates) ExecuteString(name string, base string, params any) (string, error) {
	var buffer bytes.Buffer
	if err := p.Execute(name, &buffer, base, params); err != nil {
		return "", nil
	}

	return buffer.String(), nil
}

func (p *Templates) ExecutePlainString(name string, params any) (string, error) {
	return p.ExecuteString(name, "", params)
}

func (p *Templates) executeReload(name string, w io.Writer, base string, params any) error {
	if p.dev {
		if err := p.loadOneFromDisk(name); err != nil {
			slog.Warn("error loading template from disk", "error", err)
		}
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	tmpl, ok := p.t[name]
	if !ok {
		return fmt.Errorf("template not found: %s", name)
	}

	if base == "" {
		return tmpl.Execute(w, params)
	} else {
		return tmpl.ExecuteTemplate(w, base, params)
	}
}
