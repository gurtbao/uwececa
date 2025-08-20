package web

import "net/http"

type Form map[string][]string

func (v Form) Value(s string) string {
	if vs := v[s]; len(vs) > 0 {
		return vs[0]
	}
	return ""
}

type FromForm interface {
	From(f Form) error
}

func FromMultipart[T FromForm](r *http.Request, s T) error {
	if r.Form == nil {
		// Parse with a default max memory of 32mb.
		r.ParseMultipartForm(32 << 20)
	}

	return s.From(Form(r.Form))
}
