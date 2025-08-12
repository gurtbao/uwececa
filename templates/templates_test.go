package templates_test

import (
	"bytes"
	"embed"
	"testing"

	"github.com/stretchr/testify/require"
	"uwece.ca/app/templates"
)

//go:embed templates/*
var embedFS embed.FS

func TestRenderFragment(t *testing.T) {
	t.Parallel()
	templ := templates.NewTemplates(embedFS)
	var data bytes.Buffer

	err := templ.ExecutePlain("fragments/fragment", &data, "Hallo")
	require.NoError(t, err)

	require.Equal(t, "\n<h1>Fragment: Hallo</h1>\n", data.String())
}

func TestRenderPage(t *testing.T) {
	t.Parallel()
	templ := templates.NewTemplates(embedFS)
	var data bytes.Buffer

	err := templ.Execute("home", &data, "layouts/base", nil)
	require.NoError(t, err)

	require.Equal(t, "\n<h1>Base Layout</h1>\n\n<h1>Hallo</h1>\n\n", data.String())
}
