package test

import (
	"testing"

	"github.com/johnfercher/maroto/v2"
	"github.com/parxyws/cozybox/internal/tools/pdf"
	"github.com/parxyws/cozybox/internal/tools/pdf/components"
)

func TestBuilderHeader(t *testing.T) {
	cfg := pdf.BuildConfig()
	m := maroto.New(cfg)
	m.RegisterHeader(components.BuildHeader()...)
	doc, err := m.Generate()
	if err != nil {
		t.Error(err)
	}
	doc.Save("header.pdf")
}

func TestBuilderFooter(t *testing.T) {
	cfg := pdf.BuildConfig()
	m := maroto.New(cfg)
	m.RegisterFooter(components.BuildFooter())
	doc, err := m.Generate()
	if err != nil {
		t.Error(err)
	}
	doc.Save("footer.pdf")
}

func TestBuilderDocument(t *testing.T) {
	cfg := pdf.BuildConfig()
	m := maroto.New(cfg)
	m.RegisterHeader(components.BuildHeader()...)
	m.RegisterFooter(components.BuildFooter())
	doc, err := m.Generate()
	if err != nil {
		t.Error(err)
	}
	doc.Save("document.pdf")
}
