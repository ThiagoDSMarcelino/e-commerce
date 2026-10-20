package email

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	"promotions/internal/promotions"
	"strings"
	texttemplate "text/template"
)

//go:embed templates/*
var templatesFS embed.FS

var (
	htmlTemplate = htmltemplate.Must(htmltemplate.ParseFS(templatesFS, "templates/promotion.html"))
	textTemplate = texttemplate.Must(texttemplate.New("promotion.txt").
			Funcs(texttemplate.FuncMap{"upper": strings.ToUpper}).
			ParseFS(templatesFS, "templates/promotion.txt"))
)

type Message struct {
	Subject string
	HTML    string
	Text    string
}

type promotionView struct {
	ProductID     string
	ProductName   string
	CategoryLabel string
	Discount      int
	ValidUntil    string
}

func Render(p promotions.Promotion) (Message, error) {
	view := promotionView{
		ProductID:     p.ProductID,
		ProductName:   p.ProductName,
		CategoryLabel: p.CategoryLabel,
		Discount:      p.Discount,
		ValidUntil:    p.ValidUntil.Format("02/01/2006"),
	}

	var html bytes.Buffer
	if err := htmlTemplate.Execute(&html, view); err != nil {
		return Message{}, fmt.Errorf("failed to render html: %w", err)
	}

	var text bytes.Buffer
	if err := textTemplate.Execute(&text, view); err != nil {
		return Message{}, fmt.Errorf("failed to render text: %w", err)
	}

	return Message{
		Subject: fmt.Sprintf("%d%% OFF em %s", p.Discount, p.ProductName),
		HTML:    html.String(),
		Text:    text.String(),
	}, nil
}
