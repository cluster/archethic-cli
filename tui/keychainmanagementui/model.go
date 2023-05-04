package keychainmanagementui

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/archethic-foundation/archethic-cli/tui/keychaincreatetransactionui"
	archethic "github.com/archethic-foundation/libgo"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BackMsg change state back to project view
type BackMsg bool
type SendNewKeychainTransaction struct {
	Error error
}

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	accessFocusedButton = focusedStyle.Copy().Render("[ Access Keychain ]")
	accessBlurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Access Keychain"))

	createFocusedButton = focusedStyle.Copy().Render("[ Create Keychain ]")
	createBlurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Create Keychain"))

	createTransactionFocusedButton       = focusedStyle.Copy().Render("[ Create Transaction ]")
	createTransactionaccessBlurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Create Transaction"))
	urlType                              = []string{"Local", "Testnet", "Mainnet", "Custom"}
	urls                                 = map[string]string{
		"Local":   "http://localhost:4000",
		"Testnet": "https://testnet.archethic.net",
		"Mainnet": "https://mainnet.archethic.net",
		"Custom":  ""}
)

type Model struct {
	focusIndex                       int
	inputs                           []textinput.Model
	keychain                         *archethic.Keychain
	serviceNames                     []string
	selectedUrl                      string
	selectedService                  int
	keychainSeed                     string
	keychainTransactionAddress       string
	keychainAccessTransactionAddress string
	feedback                         string
	showLoading                      bool
	spinner                          spinner.Model
}

func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	m := Model{
		inputs:  make([]textinput.Model, 2),
		spinner: s,
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.CursorStyle = cursorStyle

		switch i {
		case 0:
			t.Prompt = ""
		case 1:
			t.Prompt = "> Access seed\n"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	urlBlockSize := len(urlType)
	switch msg := msg.(type) {
	case SendNewKeychainTransaction:
		if msg.Error != nil {
			m.showLoading = false
			m.feedback = msg.Error.Error()
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			return m, func() tea.Msg {
				return BackMsg(true)
			}

		case "enter":
			if m.focusIndex < urlBlockSize {
				u := urlType[m.focusIndex]
				m.inputs[0].SetValue(urls[u])
				m.selectedUrl = u
				m.focusIndex = urlBlockSize
			}

			// create keychain button
			if m.focusIndex == len(m.inputs)+urlBlockSize {
				m.feedback = ""
				err := validateInput(m)
				if err != nil {
					m.feedback = err.Error()
					return m, nil
				}
				m.showLoading = true
				err := validateInput(m)
				if err != nil {
					m.feedback = err.Error()
					return m, nil
				}
				return m, func() tea.Msg {
					return SendNewKeychainTransaction{createKeychain(&m)}
				}
			}

			// access keychain button
			if m.focusIndex == len(m.inputs)+5 {
				err := validateInput(m)
				m.feedback = ""
				if err != nil {
					m.feedback = err.Error()
					return m, nil
				}
				url := m.inputs[0].Value()
				seed, err := archethic.MaybeConvertToHex(m.inputs[1].Value())
				if err != nil {
					m.feedback = err.Error()
					return m, nil
				}
				client := archethic.NewAPIClient(url)
				m.keychain, err = archethic.GetKeychain(seed, *client)
				if err != nil {
					m.feedback = err.Error()
					return m, nil
				}
				m.serviceNames = make([]string, 0, len(m.keychain.Services))
				for k := range m.keychain.Services {
					m.serviceNames = append(m.serviceNames, k)
				}
				sort.Strings(m.serviceNames)
			}

			// select service
			if m.focusIndex > len(m.inputs)+urlBlockSize+1 && m.focusIndex < len(m.inputs)+urlBlockSize+2+len(m.serviceNames) {
				m.selectedService = m.focusIndex - len(m.inputs) - urlBlockSize - 2
			}

			// redirect to create transaction
			if m.focusIndex == len(m.inputs)+6+len(m.serviceNames) {
				return m, func() tea.Msg {
					return keychaincreatetransactionui.CreateTransactionMsg{
						ServiceName: m.serviceNames[m.selectedService],
						Url:         m.inputs[0].Value(),
						Seed:        m.inputs[1].Value(),
					}
				}
			}

		// Set focus to next input
		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else if s == "down" || s == "tab" {
				m.focusIndex++
			}

			serviceSize := len(m.serviceNames)
			// if there is at least one service, there is a button to create a transaction
			if serviceSize > 0 {
				serviceSize++
			}
			if m.focusIndex > len(m.inputs)+urlBlockSize+1+serviceSize {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) + urlBlockSize + 1 + serviceSize
			}
		}
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Handle character input
	cmds := m.updateInputs(msg)

	cmds = append(cmds, m.updateFocus(urlBlockSize)...)

	return m, tea.Batch(cmds...)
}

func validateInput(m Model) error {
	if m.inputs[0].Value() == "" {
		return errors.New("please select a node endpoint")
	}
	if m.inputs[1].Value() == "" {
		return errors.New("please enter a seed")
	}
	return nil
}
func (m *Model) updateInputs(msg tea.Msg) []tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return cmds
}

func (m *Model) updateFocus(urlBlockSize int) []tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := 0; i <= len(m.inputs)-1; i++ {
		if i == m.focusIndex-urlBlockSize {
			// Set focused state
			cmds[i] = m.inputs[i].Focus()
			continue
		}
		// Remove focused state
		m.inputs[i].Blur()
		m.inputs[i].PromptStyle = noStyle
		m.inputs[i].TextStyle = noStyle
	}

	return cmds
}

func createKeychain(m *Model) error {
	url := m.inputs[0].Value()
	accessSeed, err := archethic.MaybeConvertToHex(m.inputs[1].Value())
	if err != nil {
		return err
	}
	originPrivateKey, _ := hex.DecodeString("01019280BDB84B8F8AEDBA205FE3552689964A5626EE2C60AA10E3BF22A91A036009")

	publicKey, _, err := archethic.DeriveKeypair(accessSeed, 0, archethic.ED25519)
	if err != nil {
		return err
	}

	randomSeed := make([]byte, 32)
	rand.Read(randomSeed)

	accessAddress, err := archethic.DeriveAddress(accessSeed, 1, archethic.ED25519, archethic.SHA256)
	if err != nil {
		return err
	}
	keychainAddress, err := archethic.DeriveAddress(randomSeed, 1, archethic.ED25519, archethic.SHA256)
	if err != nil {
		return err
	}

	keychainTx, err := archethic.NewKeychainTransaction(randomSeed, [][]byte{publicKey})
	if err != nil {
		return err
	}
	keychainTx.OriginSign(originPrivateKey)

	client := archethic.NewAPIClient(url)
	accessKeychain, _ := archethic.GetKeychain(accessSeed, *client)
	if accessKeychain != nil {
		return errors.New("keychain access already exists")
	}

	ts := archethic.NewTransactionSender(client)
	ts.AddOnRequiredConfirmation(func(nbConf int) {
		m.feedback += "\nKeychain's transaction confirmed."

		m.keychainSeed = hex.EncodeToString(randomSeed)
		m.keychainTransactionAddress = fmt.Sprintf("%s/explorer/transaction/%x", url, keychainAddress)

		accessTx, err := archethic.NewAccessTransaction(accessSeed, keychainAddress)
		if err != nil {
			m.feedback = err.Error()
		}
		accessTx.OriginSign(originPrivateKey)
		ts2 := archethic.NewTransactionSender(client)
		ts2.AddOnRequiredConfirmation(func(nbConf int) {
			m.showLoading = false
			m.feedback += "\nKeychain access transaction confirmed."
			ts2.Unsubscribe("confirmation")
			m.keychainAccessTransactionAddress = fmt.Sprintf("%s/explorer/transaction/%x", url, accessAddress)
		})
		ts2.AddOnError(func(senderContext, message string) {
			m.showLoading = false
			m.feedback += fmt.Sprintf("\nAccess transaction error: %s", message)
			ts.Unsubscribe("error")
		})
		ts2.SendTransaction(accessTx, 100, 60)
		ts.Unsubscribe("confirmation")
	})
	ts.AddOnError(func(senderContext, message string) {
		m.showLoading = false
		m.feedback += fmt.Sprintf("Keychain transaction error: %s", message)
		ts.Unsubscribe("error")
	})
	ts.SendTransaction(keychainTx, 100, 60)
	return nil
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString("> Node endpoint:\n")
	b.WriteString(urlView(m))
	for i := range m.inputs {
		b.WriteRune('\n')
		b.WriteString(m.inputs[i].View())
	}

	if m.showLoading {
		b.WriteString("\n\n")
		b.WriteString(m.spinner.View())
	}

	createButton := &createBlurredButton
	if m.focusIndex == len(m.inputs)+len(urlType) {
		createButton = &createFocusedButton
	}

	if m.feedback != "" {
		b.WriteString("\n\n")
		b.WriteString(m.feedback)
		b.WriteString("\n\n")
	}

	fmt.Fprintf(&b, "\n\n%s", *createButton)
	b.WriteRune('\n')
	if m.keychainSeed != "" {
		fmt.Fprintf(&b, "Keychain seed: %s\n", m.keychainSeed)
		fmt.Fprintf(&b, "Keychain transaction: %s\n", m.keychainTransactionAddress)
		fmt.Fprintf(&b, "Keychain access transaction: %s\n", m.keychainAccessTransactionAddress)
	}

	button := &accessBlurredButton
	if m.focusIndex == len(m.inputs)+5 {
		button = &accessFocusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	if m.keychain != nil {
		for i, k := range m.serviceNames {

			var u string
			if m.selectedService == i {
				u = "(•) "
			} else {
				u = "( ) "
			}
			u += k + " : " + m.keychain.Services[k].DerivationPath + "\n"
			if m.focusIndex == i+len(m.inputs)+6 {
				b.WriteString(focusedStyle.Render(u))
			} else {
				b.WriteString(u)
			}
			b.WriteString("\n")
		}

		button := &createTransactionaccessBlurredButton
		if m.focusIndex == len(m.inputs)+len(m.serviceNames)+6 {
			button = &createTransactionFocusedButton
		}
		fmt.Fprintf(&b, "\n\n%s\n\n", *button)
	}

	b.WriteString(helpStyle.Render("press 'esc' to go back "))

	return b.String()
}

func urlView(m Model) string {
	s := strings.Builder{}

	for i := 0; i < len(urlType); i++ {
		var u string
		if m.selectedUrl == urlType[i] {
			u = "(•) "
		} else {
			u = "( ) "
		}
		u += urlType[i]
		if i == m.focusIndex {
			s.WriteString(focusedStyle.Render(u))
		} else {
			s.WriteString(u)
		}
		s.WriteString("\n")
	}

	return s.String()
}
