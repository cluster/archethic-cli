package keychaincreatetransactionui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	CursorUp    key.Binding
	CursorDown  key.Binding
	Esc         key.Binding
	Enter       key.Binding
	ForceQuit   key.Binding
	CursorRight key.Binding
	CursorLeft  key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Browsing.
		CursorUp: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down"),
		),
		CursorLeft: key.NewBinding(
			key.WithKeys("left", "shift+tab"),
			key.WithHelp("←/shift+tab", "left"),
		),
		CursorRight: key.NewBinding(
			key.WithKeys("right", "tab"),
			key.WithHelp("→/tab", "right"),
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
