package dynamiccli

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mitchellh/go-dynamic-cli/internal/flex"
)

var debugTree = false

func tree(
	parent *flex.Node,
	c Component,
	termRows, termCols uint,
) {
	// Don't do anything with no component
	if c == nil {
		return
	}

	// Setup our node
	node := flex.NewNodeWithConfig(parent.Config)
	parent.InsertChild(node, len(parent.Children))

	// Notify of the terminal size
	if c, ok := c.(ComponentTerminalSizer); ok {
		c.SetTerminalSize(termRows, termCols)
	}

	// Setup a custom layout
	if c, ok := c.(componentLayout); ok {
		c.Layout().Apply(node)
	}

	switch c := c.(type) {
	case *fragmentComponent:
		for _, c := range c.List {
			tree(parent, c, termRows, termCols)
		}

	case *TextComponent:
		// If this is a terminal node then we setup extra styles
		node.Context = &nodeContext{
			Component: c,
		}

		node.StyleSetFlexShrink(1)
		node.StyleSetFlexGrow(0)
		node.StyleSetFlexDirection(flex.FlexDirectionRow)
		node.SetMeasureFunc(measureNode)

	default:
		// If this is not terminal then we nest.
		tree(node, c.Body(), termRows, termCols)
	}

}

func renderTree(w io.Writer, parent *flex.Node, lastRow int) {
	if debugTree {
		if w != ioutil.Discard {
			fmt.Printf("parent left: %f\n", parent.LayoutGetLeft())     // 0
			fmt.Printf("parent top: %f\n", parent.LayoutGetTop())       // 0
			fmt.Printf("parent width: %f\n", parent.LayoutGetWidth())   // 200
			fmt.Printf("parent height: %f\n", parent.LayoutGetHeight()) // 200
			defer os.Exit(1)
		}

		w = ioutil.Discard
	}

	for i, child := range parent.Children {
		// Debug. Flip this to true to see flexbox calculations.
		if debugTree {
			fmt.Printf("child %d left: %f\n", i, child.LayoutGetLeft())     // 0
			fmt.Printf("child %d top: %f\n", i, child.LayoutGetTop())       // 0
			fmt.Printf("child %d width: %f\n", i, child.LayoutGetWidth())   // 200
			fmt.Printf("child %d height: %f\n", i, child.LayoutGetHeight()) // 200
		}

		// If we're on a different row than last time then we draw a newline.
		thisRow := int(child.LayoutGetTop())
		if lastRow >= 0 && thisRow > lastRow {
			fmt.Fprintln(w)
		}
		lastRow = thisRow

		// If we have a left margin, draw that first.
		if v := int(child.LayoutGetMargin(flex.EdgeLeft)); v > 0 {
			fmt.Fprint(w, strings.Repeat(" ", v))
		}

		// Get our node context. If we don't have one then we're a container
		// and we render below.
		ctx, ok := child.Context.(*nodeContext)
		if !ok {
			renderTree(w, child, lastRow)
			continue
		}

		text := ctx.Text

		// If the height/width that the layout engine calculated is less than
		// the height that we originally measured, then we need to give the
		// element a chance to rerender into that dimension. If it still exceeds
		// it, we truncate.
		height := child.LayoutGetHeight()
		width := child.LayoutGetWidth()
		if height < ctx.Size.Height || width < ctx.Size.Width {
			// Rerender into it
			measureNode(child,
				width, flex.MeasureModeAtMost,
				height, flex.MeasureModeAtMost,
			)
			text = ctx.Text

			// Truncate, no-ops if it fits.
			text = truncateTextHeight(text, int(height))
		}

		// Draw our text
		fmt.Fprint(w, text)
	}
}

type nodeContext struct {
	Component *TextComponent
	Text      string
	Size      flex.Size
}