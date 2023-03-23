package generateaddressui

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	archethic "github.com/archethic-foundation/libgo"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BackMsg change state back to project view
type BackMsg bool

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

func GetHashAlgorithmName(h archethic.HashAlgo) string {
	switch h {
	case archethic.SHA256:
		return "SHA256"
	case archethic.SHA512:
		return "SHA512"
	case archethic.SHA3_256:
		return "SHA3_256"
	case archethic.SHA3_512:
		return "SHA3_512"
	case archethic.BLAKE2B:
		return "BLAKE2B"
	}
	panic("Unknown hash algorithm")
}

func GetCurveName(h archethic.Curve) string {
	switch h {
	case archethic.ED25519:
		return "ED25519"
	case archethic.P256:
		return "P256"
	case archethic.SECP256K1:
		return "SECP256K1"
	}
	panic("Unknown curve")
}

type Model struct {
	focusIndex       int
	inputs           []textinput.Model
	generatedAddress string
}

func New() Model {
	m := Model{
		inputs: make([]textinput.Model, 4),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.CursorStyle = cursorStyle

		switch i {
		case 0:
			t.Prompt = "> Key generation seed:\n"
			t.Focus()
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		case 1:
			t.Prompt = "> Index of key to generate\n"
			t.Placeholder = "(default 0)"
			t.Validate = numberValidator
		case 2:
			t.Prompt = "> Elliptic curve\n"
			t.Placeholder = "(default 0)"
			t.CharLimit = 1
			t.Validate = curveValidator
		case 3:
			t.Prompt = "> Hash algorithm\n"
			t.Placeholder = "(default 0)"
			t.CharLimit = 1
			t.Validate = hashAlgoValidator
		}

		m.inputs[i] = t
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func curveValidator(s string) error {
	val, err := strconv.ParseInt(s, 10, 32)
	if err == nil && (val < 0 || val > 2) {
		return errors.New("Number should be >0 and <=2")
	}
	return err
}

func hashAlgoValidator(s string) error {
	val, err := strconv.ParseInt(s, 10, 32)
	if err == nil && (val < 0 || val > 4) {
		return errors.New("Number should be >0 and <=4")
	}
	return err
}

func numberValidator(s string) error {
	_, err := strconv.ParseInt(s, 10, 32)
	return err
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			return m, func() tea.Msg {
				return BackMsg(true)
			}

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.focusIndex == len(m.inputs) {
				seed := m.inputs[0].Value()
				index, err := strconv.ParseUint(m.inputs[1].Value(), 10, 32)

				// check for errors
				if err != nil {
					index = 0
				}
				curveInt, err := strconv.ParseUint(m.inputs[2].Value(), 10, 8)
				if err != nil {
					curveInt = 0
				}
				hashAlgoInt, err := strconv.ParseUint(m.inputs[3].Value(), 10, 8)
				if err != nil {
					hashAlgoInt = 0
				}

				curve := archethic.Curve(uint8(curveInt))
				hashAlgo := archethic.HashAlgo(uint8(hashAlgoInt))

				address := archethic.DeriveAddress([]byte(seed), uint32(index), curve, hashAlgo)
				m.generatedAddress = hex.EncodeToString(address)
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					// m.inputs[i].PromptStyle = focusedStyle
					// m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m Model) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i == 2 {
			b.WriteString("\n")
			for j := 0; j <= 2; j++ {
				b.WriteString("\t (" + strconv.Itoa(j) + ") " + GetCurveName(archethic.Curve(j)) + "\n")
			}
		}
		if i == 3 {
			b.WriteString("\n")
			for j := 0; j <= 4; j++ {
				b.WriteString("\t (" + strconv.Itoa(j) + ") " + GetHashAlgorithmName(archethic.HashAlgo(j)) + "\n")
			}
		}
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	if m.generatedAddress != "" {
		b.WriteString("The generated address is: " + m.generatedAddress)
	}

	b.WriteString(helpStyle.Render("press 'esc' to go back "))

	return b.String()
}
