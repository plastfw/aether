package ui

import "strings"

type keyHintItem struct {
	Keys        []string
	Description string
}

func keyHint(keys []string, description string) keyHintItem {
	return keyHintItem{Keys: keys, Description: description}
}

func (m Model) renderKeyHint(hint keyHintItem) string {
	return consoleHotStyle.Render(m.displayKeys(hint.Keys)) + consoleDimStyle.Render(" "+hint.Description)
}

func renderKeyHint(hint keyHintItem) string {
	return consoleHotStyle.Render(displayKeys(hint.Keys)) + consoleDimStyle.Render(" "+hint.Description)
}

func renderKeyHintLine(width int, hints ...keyHintItem) string {
	return fitLine(renderKeyHintText(hints...), width)
}

func renderKeyHintText(hints ...keyHintItem) string {
	parts := make([]string, 0, len(hints))
	for _, hint := range hints {
		if len(hint.Keys) == 0 {
			continue
		}
		parts = append(parts, renderKeyHint(hint))
	}
	return strings.Join(parts, consoleDimStyle.Render("  •  "))
}

func (m Model) renderKeyHelpLine(width int, hint keyHintItem) string {
	return fitLine(consoleHotStyle.Render(m.displayKeys(hint.Keys)), 12) + consoleDimStyle.Render(hint.Description)
}

func renderKeyHelpLine(width int, hint keyHintItem) string {
	return fitLine(consoleHotStyle.Render(displayKeys(hint.Keys)), 12) + consoleDimStyle.Render(hint.Description)
}

func (m Model) displayKeys(keys []string) string {
	labels := make([]string, 0, len(keys))
	used := map[string]bool{}
	for _, key := range keys {
		label := m.displayKey(key)
		if used[label] {
			continue
		}
		used[label] = true
		labels = append(labels, label)
	}
	return strings.Join(labels, " / ")
}

func (m Model) displayKey(key string) string {
	for _, alias := range m.settings.UI.KeyAliases {
		for _, aliasKey := range alias.Keys {
			if key == aliasKey && alias.Label != "" {
				return alias.Label
			}
		}
	}
	return displayKey(key)
}

func displayKeys(keys []string) string {
	labels := make([]string, 0, len(keys))
	for _, key := range keys {
		labels = append(labels, displayKey(key))
	}
	return strings.Join(labels, " / ")
}

func displayKey(key string) string {
	switch key {
	case "ctrl+c":
		return "Ctrl+C"
	case "enter":
		return "Enter"
	case "tab":
		return "Tab"
	case "up":
		return "↑"
	case "down":
		return "↓"
	case "left":
		return "←"
	case "right":
		return "→"
	case "delete":
		return "Del"
	case " ":
		return "Space"
	default:
		return key
	}
}
