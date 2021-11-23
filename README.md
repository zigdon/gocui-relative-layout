# gocui-relative-layout

This module provides a layout manager, appropriate for using with gocui's
SetManager function. The manager builds a full screen layout, dividing between
multiple items with the ratios or sizes provided.

## Example

	  +-------------------------------+
	  | view 1                        |
	  +-------------------------------+
	  | view 2 |  view 3              |
	  |        +----------------------+
	  |        |  view 4              |
	  |        +----------------------+
	  |        |  view 5              |
	  +-------------------------------+

```
  import (
	"github.com/awesome-gocui/gocui"
    "github.com/zigdon/gocui-relative-layout rl"
  )

  viewCol := rl.NewLevel(
    rl.LayoutVertical,
    rl.NewRatioItem(1, "view3", nil),                       # 33% of the space
    rl.NewRatioItem(1, "view4", nil),                       # 33% of the space
    rl.NewRatioItem(1, "view5", nil),                       # 33% of the space
  )

  content := rl.NewLevel(
    rl.LayoutHorizontal,
    rl.NewFixedItem(15, "view2", nil),                      # <--- 15 columns
    rl.NewRatioItem(1, "_viewCol", rl.WithInner(viewCol)),  # <--- remainin space
  )

  layout := rl.NewLevel(
    rl.LayoutVertical,
    rl.NewRatioItem(1, "view1", nil),                       # <--- 25% of the screen
    rl.NewRatioItem(3, "_content", rl.WithInner(content)),  # <--- 75% of the screen
  )

  gui := gocui.NewGui(gocui.OutputNormal, true)
  gui.SetManager(layout)
```

## Creating Items

To create new items for the layout use the `NewFixedItem` or `NewRatioItem`
functions. An "FixedItem" has a set number of lines in the UI. Any lines not
used up by FixedItems are distributed between all the remainin RatioItems,
weighted by each's item size.

### Item Options

* Hidden() - Create the view, but don't render it on screen
* WithCreate() - Call the provided function after creating the new. Useful for
  setting additional attributes on the view.
* WithUpdate() - Call the provided functoin each time the layout is rendered.
* WithInner() - This item contains additional layout items, rather than gocui Views.

## Hiding Items

Any section of the layout can be hidden. A hidden item still exists, the views
within it are still rendered and their create/update functions are still
called. Their contents, however, is not visible on the screen, instead it is
rendered "below" the other visible views.

To change the visible of an item, call `layout.HideItem(name, visibility)`,
where the visiblity is one of `LayoutHidden` or `LayoutVisible`. Alternatively,
use `layout.ToggleItem(name)` to switch between visible and hidden states.
