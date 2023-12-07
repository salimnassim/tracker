package tracker

import (
	"fmt"
	"html/template"
	"io"
	"sync"
)

type TemplateID int

const (
	TemplateIndex = iota
	TemplateTorrent
)

type Templater interface {
	Add(ID TemplateID, template *template.Template)
	Execute(ID TemplateID, wr io.Writer, data any) error
}

type templateStore struct {
	mu        *sync.RWMutex
	templates map[TemplateID]*template.Template
}

func NewTemplateStore() *templateStore {
	return &templateStore{
		mu:        &sync.RWMutex{},
		templates: make(map[TemplateID]*template.Template),
	}
}

func (c *templateStore) Add(ID TemplateID, template *template.Template) {
	c.mu.Lock()
	c.templates[ID] = template
	c.mu.Unlock()
}

func (c *templateStore) Execute(ID TemplateID, wr io.Writer, data any) error {
	c.mu.RLock()
	tpl, ok := c.templates[ID]
	c.mu.RUnlock()
	if !ok {
		return fmt.Errorf("template %q does not exist in cache", ID)
	}
	tpl.Execute(wr, data)
	return nil
}
