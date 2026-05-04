package ui

type layoutMode int

const (
	layoutTooSmall layoutMode = iota
	layoutMedium
	layoutWide
)

const (
	minTerminalWidth  = 80
	minTerminalHeight = 16
)

type layoutSpec struct {
	mode   layoutMode
	width  int
	height int
	innerW int
	innerH int
}

func computeLayout(width, height int) layoutSpec {
	spec := layoutSpec{mode: layoutTooSmall, width: width, height: height, innerW: max(1, width-2), innerH: max(1, height-2)}
	if width < minTerminalWidth || height < minTerminalHeight {
		return spec
	}
	if width >= 104 && height >= 22 {
		spec.mode = layoutWide
		return spec
	}
	spec.mode = layoutMedium
	return spec
}
