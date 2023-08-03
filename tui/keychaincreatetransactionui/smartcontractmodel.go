package keychaincreatetransactionui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type SmartContractModel struct {
	smartContractTextAreaInput textarea.Model
}

type UpdateSmartContract struct {
	Code string
}

func NewSmartContractModel() SmartContractModel {
	m := SmartContractModel{}
	m.smartContractTextAreaInput = textarea.New()
	m.smartContractTextAreaInput.CharLimit = 0
	m.smartContractTextAreaInput.SetWidth(150)
	return m
}

func (m SmartContractModel) Init() tea.Cmd {
	return nil
}

func (m SmartContractModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "esc":

			if m.smartContractTextAreaInput.Focused() {
				m.smartContractTextAreaInput.Blur()
			}
			return m, nil
		// this is used to get a faster paste
		case "ctrl+v":

			if !m.smartContractTextAreaInput.Focused() {
				m.smartContractTextAreaInput.Focus()
			}
			newText := textarea.Paste()
			m.smartContractTextAreaInput, _ = m.smartContractTextAreaInput.Update(newText)
			return m, func() tea.Msg {
				return UpdateSmartContract{Code: m.smartContractTextAreaInput.Value()}
			}
		default:

			if !m.smartContractTextAreaInput.Focused() {
				m.smartContractTextAreaInput.Focus()
			}
			m.smartContractTextAreaInput, _ = m.smartContractTextAreaInput.Update(msg)
			return m, func() tea.Msg {
				return UpdateSmartContract{Code: m.smartContractTextAreaInput.Value()}
			}

		}

	}

	return m, nil
}

func (m *SmartContractModel) SwitchTab() (SmartContractModel, []tea.Cmd) {
	return *m, nil
}

func (m SmartContractModel) View() string {
	var b strings.Builder

	b.WriteString(m.smartContractTextAreaInput.View())
	if m.smartContractTextAreaInput.Focused() {
		b.WriteString(helpStyle.Render("\npress 'esc' to exit edit mode "))
	}
	return b.String()
}
