package keychainmanagementui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	CursorUp   key.Binding
	CursorDown key.Binding
	Esc        key.Binding
	Enter      key.Binding
	ForceQuit  key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Browsing.
		CursorUp: key.NewBinding(
			key.WithKeys("up", "shift+tab"),
			key.WithHelp("↑/shift+tab", "up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down", "tab"),
			key.WithHelp("↓/tab", "down"),
		),

		// Selecting.
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "enter"),
		),

		// Quitting.
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),

		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}
