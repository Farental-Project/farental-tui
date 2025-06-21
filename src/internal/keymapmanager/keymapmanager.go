package keymapmanager

import (
	"github.com/charmbracelet/lipgloss"
	"math"
	"strings"
)

type KeymapManager struct {
	CurrentContext KeymapContext
	Contexts       map[KeymapContext]*Keymap
	ShowAll        bool
}

func NewKeymapManager() *KeymapManager {
	return &KeymapManager{
		Contexts: make(map[KeymapContext]*Keymap),
	}
}

func (m *KeymapManager) RegisterContext(context KeymapContext, keymap *Keymap) {
	m.Contexts[context] = keymap
}

func (m *KeymapManager) GetCurrentContextKeymap() *Keymap {
	ctx, ok := m.Contexts[m.CurrentContext]

	if !ok {
		return nil
	}

	return ctx
}

func (m *KeymapManager) SwitchContext(context KeymapContext) {
	if _, ok := m.Contexts[context]; ok {
		m.CurrentContext = context
		m.ShowAll = false
	}
}

func (m *KeymapManager) View(width int) string {
	keymap := m.GetCurrentContextKeymap()

	if keymap == nil {
		return "ERROR : UNKNOWN KEYMAP CONTEXT"
	}

	if m.ShowAll {
		return m.ViewAll(keymap, width)
	} else {
		return m.ViewEssential(keymap, width)
	}
}

func (m *KeymapManager) ViewAll(keymap *Keymap, width int) string {
	var keys []Key
	var columns []string
	var keyStr, sepStr, descStr strings.Builder

	keys = keymap.AllBindings()

	colCount := keymap.ShowAllColumnCount
	rowCount := int(math.Ceil(float64(len(keys)) / float64(colCount)))

	columns = make([]string, 0)

	for i, key := range keys {
		remainingCount := len(keys) - i
		notLastCol := len(columns)+1 < colCount

		if i%rowCount > 0 || (remainingCount == 1 && notLastCol) {
			keyStr.WriteString("\n")
			sepStr.WriteString("\n")
			descStr.WriteString("\n")
		}

		keyStr.WriteString(keymap.Style.FullKey.
			Render(key.Binding.Help().Key))
		sepStr.WriteString(keymap.Style.FullKeySeparator.
			Render(keymap.Style.FullKeySeparatorValue))
		descStr.WriteString(keymap.Style.FullKeyDescription.
			Render(key.Binding.Help().Desc))

		if ((i+1)%rowCount == 0 && i != 0 && notLastCol) || (remainingCount == 1) {

			columns = append(columns, lipgloss.
				JoinHorizontal(lipgloss.Center,
					keyStr.String(),
					sepStr.String(),
					descStr.String()))

			if i < len(keys)-1 {
				columns = append(columns, keymap.Style.FullColSeparatorValue)
				colCount++
			}

			keyStr.Reset()
			sepStr.Reset()
			descStr.Reset()
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top,
		columns...)
}

func (m *KeymapManager) ViewEssential(keymap *Keymap, width int) string {
	var b strings.Builder
	var keys []Key

	keys = keymap.EssentialBindings()

	for i, key := range keys {
		if i > 0 {
			b.WriteString(keymap.Style.EssentialColSeparator.
				Render(keymap.Style.EssentialColSeparatorValue))
		}

		b.WriteString(keymap.Style.EssentialKey.
			Render(key.Binding.Help().Key))

		b.WriteString(keymap.Style.EssentialKeySeparator.
			Render(keymap.Style.EssentialKeySeparatorValue))

		b.WriteString(keymap.Style.EssentialKeyDescription.
			Render(key.Binding.Help().Desc))
	}

	return lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Center).
		Width(width).Render(b.String())
}
