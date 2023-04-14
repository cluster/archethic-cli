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
	// passing 0 or a negative number here doesn't seem to work...
	m.smartContractTextAreaInput.CharLimit = 100000000000
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
		default:

			m.smartContractTextAreaInput, _ = m.smartContractTextAreaInput.Update(msg)
			return m, func() tea.Msg {
				return UpdateSmartContract{Code: m.smartContractTextAreaInput.Value()}
			}

		}

	}

	m, cmds := updateSmartContractFocus(m)
	cmds = append(cmds, m.updateSmartContractInputs(msg)...)
	return m, tea.Batch(cmds...)
}

func (m *SmartContractModel) updateSmartContractInputs(msg tea.Msg) []tea.Cmd {

	cmds := make([]tea.Cmd, 1)
	m.smartContractTextAreaInput, cmds[0] = m.smartContractTextAreaInput.Update(msg)

	return cmds
}

func updateSmartContractFocus(m SmartContractModel) (SmartContractModel, []tea.Cmd) {

	cmds := make([]tea.Cmd, 1)
	cmds[0] = m.smartContractTextAreaInput.Focus()

	return m, cmds
}

func (m *SmartContractModel) SwitchTab() (SmartContractModel, []tea.Cmd) {
	m.smartContractTextAreaInput.Focus()
	m2, cmds := updateSmartContractFocus(*m)
	return m2, cmds
}

func (m SmartContractModel) View() string {
	var b strings.Builder

	b.WriteString(m.smartContractTextAreaInput.View())
	if m.smartContractTextAreaInput.Focused() {
		b.WriteString(helpStyle.Render("\npress 'esc' to exit edit mode "))
	}
	return b.String()
}
