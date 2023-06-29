package internal

import (
	"fmt"
	c "github.com/jroimartin/gocui"
)

type TransformFunc[T any] func(T) string
type SortOnFunc[T any] func(T) string

type BaseView struct {
	View           *c.View
	StartX, StartY int
	EndX, EndY     int
	Title          string
}

func (w *BaseView) RenderView(g *c.Gui, sx int, sy int, ex int, ey int) error {
	w.StartX, w.StartY, w.EndX, w.EndY = sx, sy, ex, ey
	_, err := g.SetView(w.View.Name(), w.StartX, w.StartY, w.EndX, w.EndY)
	w.View.Title = w.Title
	return err
}

func (w *BaseView) SetNavigableFeedback(g *c.Gui) {
	w.View.SelFgColor = c.AttrBold
	w.View.SelBgColor = c.ColorRed
	w.View.Wrap = true
}

func (w *BaseView) SetView(g *c.Gui, viewName string) error {
	tempView, err := g.SetView(viewName, 0, 0, 1, 1)
	if err != nil {
		if err != c.ErrUnknownView {
			return err
		}
	}
	w.View = tempView
	return nil
}

type InfoView struct {
	Base    *BaseView
	Content []string
}

func (w *InfoView) Initialize(g *c.Gui, title string, name string) {
	w.Base = &BaseView{
		Title: title,
	}
	w.Base.SetView(g, name)
}

func (w *InfoView) RenderContent(content []string) {
	w.Content = content
	w.Base.View.Clear()
	for _, item := range w.Content {
		fmt.Fprintln(w.Base.View, item)
	}
}

type ListView[T any] struct {
	Base     *BaseView
	Elements []string
	// TODO do we really need the RawElements
	RawElements       []T
	TransformRawToStr TransformFunc[T]
	// TODO how to make string more generic, right now we have the assumption that the rendered elements will be a string
	SortElementsOn SortOnFunc[T]
}

func (w *ListView[T]) Initialize(g *c.Gui, title string, name string, transform TransformFunc[T], sortOn SortOnFunc[T]) {
	w.Base = &BaseView{
		Title: title,
	}
	w.TransformRawToStr = transform
	w.SortElementsOn = sortOn
	w.Base.SetView(g, name)
}

func (w *ListView[T]) ResetCursor() {
	w.Base.View.SetOrigin(0, 0)
	w.Base.View.SetCursor(0, 0)
}

func (w *ListView[T]) GetElementOnCursorPosition() string {
	_, oy := w.Base.View.Origin()
	_, vy := w.Base.View.Cursor()

	return w.Elements[vy+oy]
}

func (w *ListView[T]) RenderItems(items []T, sort ...bool) {
	// default sort: true
	w.RawElements = make([]T, 0)
	w.Elements = make([]string, 0)
	if len(sort) > 0 && !sort[0] {
		w.RawElements = items
	} else {
		w.RawElements = SortBy(items, w.SortElementsOn)
	}

	for _, item := range w.RawElements {
		itemStr := w.TransformRawToStr(item)
		w.Elements = append(w.Elements, itemStr)
	}
	w.Base.View.Clear()
	for _, item := range w.Elements {
		fmt.Fprintln(w.Base.View, item)
	}
}
