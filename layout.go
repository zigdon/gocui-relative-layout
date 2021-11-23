package layout

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
)

// LayoutDirection controls if layout items are spread horizontally or
// vertically.
type LayoutDirection bool

// HideLayout identifies if an item should be hidden or visible.
type HideLayout bool

// NotFound is an error returned when an item referenced by name does not
// exist.
var NotFound = fmt.Errorf("Item not found")

// InvalidValues is an error returned when both fixes and ratios are specified
// for the same item.
var InvalidValues = fmt.Errorf("Fixes and Ratio parameters are not compatible")

const (
	LayoutHorizontal LayoutDirection = true
	LayoutVertical   LayoutDirection = false

	LayoutHidden  HideLayout = true
	LayoutVisible HideLayout = false
)

type layoutItem struct {
	ratio   int
	fixed   int
	name    string
	hidden  HideLayout
	inner   *layoutLevel
	fNew    func(*gocui.View) error
	fUpdate func(*gocui.View) error
}

type layoutItemOption func(l *layoutItem)

// Hidden is an item option indicating a given item should not be visible.
func Hidden() layoutItemOption {
	return func(l *layoutItem) {
		l.hidden = true
	}
}

// WithInner is used when an item is to contain other items, rather than views.
func WithInner(inner *layoutLevel) layoutItemOption {
	return func(l *layoutItem) {
		l.inner = inner
	}
}

// WithCreate passes a function that will be called once a new view is created.
// It is called only once per view.
func WithCreate(f func(*gocui.View) error) layoutItemOption {
	return func(l *layoutItem) {
		l.fNew = f
	}
}

// WithUpdate passes a function that is called each time the layout is
// rendered.
func WithUpdate(f func(*gocui.View) error) layoutItemOption {
	return func(l *layoutItem) {
		l.fUpdate = f
	}
}

// NewRatioItem creates a new item, that is to take a given ratio of the total
// available space.
func NewRatioItem(weight int, name string, opts ...layoutItemOption) *layoutItem {
	return createNewItem(weight, name, opts...)
}

// NewFixedItem create a new item with the specified number of lines/columns.
func NewFixedItem(size int, name string, opts ...layoutItemOption) *layoutItem {
	return createNewItem(-size, name, opts...)
}

func createNewItem(size int, name string, opts ...layoutItemOption) *layoutItem {
	var ratio, fixed int
	if size > 0 {
		ratio = size
	} else if size < 0 {
		fixed = -size
	} else {
		panic("invalid size when creating layoutItem")
	}

	i := &layoutItem{
		ratio: ratio,
		fixed: fixed,
		name:  name,
	}

	for _, o := range opts {
		o(i)
	}

	return i
}

func (l *layoutItem) isHidden() HideLayout {
	if l.hidden {
		return LayoutHidden
	}
	if l.inner != nil {
		return l.inner.allHidden()
	}

	return LayoutVisible
}

type layoutLevel struct {
	direction LayoutDirection
	items     []*layoutItem
}

// NewLevel create a new set of items to be spread either horizontally or
// vertically.
func NewLevel(direction LayoutDirection, items ...*layoutItem) *layoutLevel {
	return &layoutLevel{direction, items}
}

func (l *layoutLevel) findItem(name string) (*layoutItem, error) {
	for _, item := range l.items {
		if item.name == name {
			return item, nil
		}
		if item.inner != nil {
			found, err := item.inner.findItem(name)
			if err == nil {
				return found, err
			}
			if err != NotFound {
				return nil, err
			}
		}
	}
	return nil, NotFound
}

// ToggleItem finds the item with the specified name within the layout (or
// sublayouts), and toggles its visibility.
func (l *layoutLevel) ToggleItem(name string) error {
	i, err := l.findItem(name)
	if err != nil {
		return err
	}

	i.hidden = !i.hidden

	return nil
}

// ToggleItem finds the item with the specified name within the layout (or
// sublayouts), and sets its visibility to the provided value.
func (l *layoutLevel) HideItem(name string, hidden HideLayout) error {
	i, err := l.findItem(name)
	if err != nil {
		return err
	}

	i.hidden = hidden

	return nil
}

// ResizeItem changes the space allocated for the named item. Either fixed or
// ratio values can be provided, but not both.
func (l *layoutLevel) ResizeItem(name string, ratio, fixed int) error {
	i, err := l.findItem(name)
	if err != nil {
		return err
	}

	if ratio != 0 && fixed != 0 {
		return InvalidValues
	}

	i.ratio = ratio
	i.fixed = fixed

	return nil
}

func (l *layoutLevel) allHidden() HideLayout {
	for _, item := range l.items {
		if !item.isHidden() {
			return LayoutVisible
		}
	}
	return LayoutHidden
}

func (l *layoutLevel) layout(g *gocui.Gui, x0, y0, x1, y1 int, forceHidden HideLayout) error {
	var length, acc int
	var overlap int
	if !g.SupportOverlaps {
		overlap = 1
	}

	// Figure out which dimention we care about
	if l.direction == LayoutHorizontal {
		length = x1 - x0 + 1
		acc = x0
	} else {
		length = y1 - y0 + 1
		acc = y0
	}

	// Add up all the (visible) fixed sizes, as they're not available for assignment
	fixed := 0
	segments := 0
	lastVisible := 0
	for i, item := range l.items {
		if forceHidden || item.isHidden() {
			continue
		}
		if item.fixed > 0 {
			fixed += item.fixed
		} else {
			segments += item.ratio
		}
		lastVisible = i
	}
	if length < fixed {
		return fmt.Errorf("window too small for fixed sizes: %d < %d", length, fixed)
	}
	length -= fixed

	// The rest of the space gets split between the segments
	unit := -1
	left := -1
	if segments > 0 {
		unit = length / segments
		left = length % segments
	}

	if unit == 0 {
		return fmt.Errorf("window too small for allocated units: length=%d, segments=%d", length, segments)
	}

	for idx, item := range l.items {
		// Make sure we still create all the views, even if they're not visible
		var err error
		if forceHidden || item.isHidden() {
			if item.inner != nil {
				err = item.inner.layout(g, x0, y0, x1, y1, LayoutHidden)
			} else {
				err = createView(g, item.name, x0, y0, x1, y1, 0, item.fNew, item.fUpdate)
				g.SetViewOnBottom(item.name)
			}
			if err != nil {
				return fmt.Errorf("error creating layout: %v", err)
			}
			continue
		}

		var assignment int
		if item.fixed == 0 {
			assignment = unit * item.ratio
		} else {
			assignment = item.fixed
		}

		// The last item gets the leftovers
		if idx == lastVisible {
			assignment += left
		}

		ix0, ix1, iy0, iy1 := x0, x1, y0, y1
		if l.direction == LayoutHorizontal {
			ix0 = acc
			ix1 = acc + assignment - overlap
			if ix1 > x1 {
				ix1 = x1
			}
		} else {
			iy0 = acc
			iy1 = acc + assignment - overlap
			if iy1 > y1 {
				iy1 = y1
			}
		}
		acc += assignment

		if item.inner != nil {
			err = item.inner.layout(g, ix0, iy0, ix1, iy1, LayoutVisible)
		} else {
			err = createView(g, item.name, ix0, iy0, ix1, iy1, 0, item.fNew, item.fUpdate)
		}

		if err != nil {
			return fmt.Errorf("error creating layout: %v", err)
		}
	}

	return nil
}

func (l *layoutLevel) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	return l.layout(g, 0, 0, maxX-1, maxY-1, LayoutVisible)
}

func createView(g *gocui.Gui, name string, x0, y0, x1, y1 int, overlaps byte,
	fNew func(*gocui.View) error,
	fUpdate func(*gocui.View) error) error {
	if v, err := g.SetView(name, x0, y0, x1, y1, overlaps); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		v.Autoscroll = false
		if fNew != nil {
			return fNew(v)
		}
	} else if fUpdate != nil {
		return fUpdate(v)
	}
	return nil
}
