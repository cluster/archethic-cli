package keychainmanagementui

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/archethic-foundation/archethic-cli/tui/keychaincreatetransactionui"
	archethic "github.com/archethic-foundation/libgo"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BackMsg change state back to project view
type BackMsg bool
type SendNewKeychainTransaction struct{}

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
}

func New() Model {
	m := Model{
		inputs: make([]textinput.Model, 2),
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
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	urlBlockSize := len(urlType)
	switch msg := msg.(type) {
	case SendNewKeychainTransaction:
		createKeychain(&m)
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
				m.showLoading = true
				return m, func() tea.Msg {
					return SendNewKeychainTransaction{}
				}
			}

			// access keychain button
			if m.focusIndex == len(m.inputs)+5 {
				url := m.inputs[0].Value()
				seed := archethic.MaybeConvertToHex(m.inputs[1].Value())
				client := archethic.NewAPIClient(url)
				m.keychain = archethic.GetKeychain(seed, *client)
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
	}

	// Handle character input
	cmds := m.updateInputs(msg)

	cmds = append(cmds, m.updateFocus(urlBlockSize)...)

	return m, tea.Batch(cmds...)
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

func createKeychain(m *Model) {
	url := m.inputs[0].Value()
	accessSeed := archethic.MaybeConvertToHex(m.inputs[1].Value())
	originPrivateKey, _ := hex.DecodeString("01019280BDB84B8F8AEDBA205FE3552689964A5626EE2C60AA10E3BF22A91A036009")

	publicKey, _ := archethic.DeriveKeypair(accessSeed, 0, archethic.ED25519)

	randomSeed := make([]byte, 32)
	rand.Read(randomSeed)

	accessAddress := archethic.DeriveAddress(accessSeed, 1, archethic.ED25519, archethic.SHA256)
	keychainAddress := archethic.DeriveAddress(randomSeed, 1, archethic.ED25519, archethic.SHA256)

	keychainTx := archethic.NewKeychainTransaction(randomSeed, [][]byte{publicKey})
	keychainTx.OriginSign(originPrivateKey)

	client := archethic.NewAPIClient(url)
	ts := archethic.NewTransactionSender(client)
	ts.AddOnRequiredConfirmation(func(nbConf int) {
		m.feedback += "\nKeychain's transaction confirmed."

		m.keychainSeed = hex.EncodeToString(randomSeed)
		m.keychainTransactionAddress = fmt.Sprintf("%s/explorer/transaction/%x", url, keychainAddress)

		accessTx := archethic.NewAccessTransaction(accessSeed, keychainAddress)
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
		ts2.SendTransaction(&accessTx, 100, 60)
		ts.Unsubscribe("confirmation")
	})
	ts.AddOnError(func(senderContext, message string) {
		m.showLoading = false
		m.feedback += fmt.Sprintf("Keychain transaction error: %s", message)
		ts.Unsubscribe("error")
	})
	ts.SendTransaction(&keychainTx, 100, 60)
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString("> Node endpoint:\n")
	b.WriteString(urlView(m))
	for i := range m.inputs {
		b.WriteRune('\n')
		b.WriteString(m.inputs[i].View())
	}

	createButton := &createBlurredButton
	if m.focusIndex == len(m.inputs)+len(urlType) {
		createButton = &createFocusedButton
	}
	if m.showLoading {
		b.WriteString("\n\n")
		b.WriteString("Loading...")
		b.WriteString("\n\n")
	}

	fmt.Fprintf(&b, "\n\n%s", *createButton)

	fmt.Fprintf(&b, "%s\n", m.feedback)
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
