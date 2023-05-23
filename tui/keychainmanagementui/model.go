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
	Model Model
}
type SendAccessKeychain struct {
	Model Model
}
type SendRemoveService struct {
	Model Model
}
type SendCreateService struct {
	Model Model
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

	createTransactionFocusedButton       = focusedStyle.Copy().Render("[ Create Transaction for Service ]")
	createTransactionaccessBlurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Create Transaction for Service"))
	createServiceFocusedButton           = focusedStyle.Copy().Render("[ Create Service ]")
	createServiceBlurredButton           = fmt.Sprintf("[ %s ]", blurredStyle.Render("Create Service"))
	urlType                              = []string{"Local", "Testnet", "Mainnet", "Custom"}
	urls                                 = map[string]string{
		"Local":   "http://localhost:4000",
		"Testnet": "https://testnet.archethic.net",
		"Mainnet": "https://mainnet.archethic.net",
		"Custom":  ""}
)

type Model struct {
	IsInit                           bool
	focusIndex                       int
	inputs                           []textinput.Model
	newServiceInputs                 []textinput.Model
	keychain                         *archethic.Keychain
	serviceNames                     []string
	selectedUrl                      string
	selectedService                  int
	keychainSeed                     string
	keychainTransactionAddress       string
	keychainAccessTransactionAddress string
	feedback                         string
	showSpinnerCreate                bool
	showSpinnerAccess                bool
	showSpinnerDeleteService         bool
	showSpinnerCreateService         bool
	Spinner                          spinner.Model
}

func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	m := Model{
		inputs:           make([]textinput.Model, 2),
		Spinner:          s,
		newServiceInputs: make([]textinput.Model, 2),
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

	for i := range m.newServiceInputs {
		t = textinput.New()
		t.CursorStyle = cursorStyle

		switch i {
		case 0:
			t.Prompt = "> Service name\n"
		case 1:
			t.Prompt = "> Derivation path\n"
		}

		m.newServiceInputs[i] = t
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return m.Spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	urlBlockSize := len(urlType)
	switch msg := msg.(type) {
	case SendNewKeychainTransaction:
		m.feedback = msg.Model.feedback
		m.keychainSeed = msg.Model.keychainSeed
		m.keychainTransactionAddress = msg.Model.keychainTransactionAddress
		m.keychainAccessTransactionAddress = msg.Model.keychainAccessTransactionAddress
		m.showSpinnerCreate = false
		return m, nil
	case SendAccessKeychain:
		m.keychain = msg.Model.keychain
		m.feedback = msg.Model.feedback
		m.serviceNames = msg.Model.serviceNames
		m.selectedService = msg.Model.selectedService
		m.showSpinnerAccess = false
		return m, nil
	case SendRemoveService:
		m.keychain = msg.Model.keychain
		m.feedback = msg.Model.feedback
		m.serviceNames = msg.Model.serviceNames
		m.selectedService = msg.Model.selectedService
		m.showSpinnerDeleteService = false
		return m, nil
	case SendCreateService:
		m.keychain = msg.Model.keychain
		m.feedback = msg.Model.feedback
		m.serviceNames = msg.Model.serviceNames
		m.selectedService = msg.Model.selectedService
		m.showSpinnerCreateService = false
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
				m.showSpinnerCreate = true
				err = validateInput(m)
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
				m.showSpinnerAccess = true
				return m, func() tea.Msg {
					return SendAccessKeychain{accessKeychain(&m)}
				}
			}

			// add service
			if m.focusIndex == len(m.inputs)+6+len(m.serviceNames)+1+len(m.newServiceInputs) {
				m.showSpinnerCreateService = true
				return m, func() tea.Msg {
					return SendCreateService{addService(&m)}
				}
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
			newServiceFormSize := 0
			// if there is a keychain, there is a form to add a new service (and a button)
			if m.keychain != nil {
				newServiceFormSize = len(m.newServiceInputs) + 1
			}
			if m.focusIndex > len(m.inputs)+urlBlockSize+2+serviceSize+newServiceFormSize {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) + urlBlockSize + 2 + serviceSize + newServiceFormSize
			}
		default:
			// remove the highlighted service
			if msg.String() == "d" && m.focusIndex > len(m.inputs)+urlBlockSize+1 && m.focusIndex < len(m.inputs)+urlBlockSize+2+len(m.serviceNames) {
				selectedService := m.focusIndex - len(m.inputs) - urlBlockSize - 2
				m.showSpinnerDeleteService = true
				return m, func() tea.Msg {
					return SendRemoveService{removeServiceAndRefresh(&m, selectedService)}
				}
			}
			// set a default derivation path
			if m.focusIndex == len(m.inputs)+6+len(m.serviceNames)+1 {
				serviceName := m.newServiceInputs[0].Value()
				derivationPath := "m/650'/" + serviceName + msg.String() + "/0"
				m.newServiceInputs[1].SetValue(derivationPath)
			}
		}
	default:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}

	// Handle character input
	cmds := m.updateInputs(msg)

	cmds = append(cmds, m.updateFocus(urlBlockSize)...)

	return m, tea.Batch(cmds...)
}

func addService(m *Model) Model {
	addServiceToKeychain(m, []byte(m.inputs[1].Value()), *archethic.NewAPIClient(m.inputs[0].Value()), m.newServiceInputs[0].Value(), m.newServiceInputs[1].Value())
	m.newServiceInputs[0].SetValue("")
	m.newServiceInputs[1].SetValue("")
	m.focusIndex++
	return accessKeychain(m)
}

func removeServiceAndRefresh(m *Model, selectedService int) Model {
	removeServiceFromKeychain(m, []byte(m.inputs[1].Value()), *archethic.NewAPIClient(m.inputs[0].Value()), m.serviceNames[selectedService])
	return accessKeychain(m)
}

func accessKeychain(m *Model) Model {
	err := validateInput(*m)
	m.feedback = ""
	if err != nil {
		m.feedback = err.Error()
		return *m
	}
	url := m.inputs[0].Value()
	seed, err := archethic.MaybeConvertToHex(m.inputs[1].Value())
	if err != nil {
		m.feedback = err.Error()
		return *m
	}
	client := archethic.NewAPIClient(url)
	m.keychain, err = archethic.GetKeychain(seed, *client)
	if err != nil {
		m.feedback = err.Error()
		return *m
	}
	m.serviceNames = make([]string, 0, len(m.keychain.Services))
	for k := range m.keychain.Services {
		m.serviceNames = append(m.serviceNames, k)
	}
	sort.Strings(m.serviceNames)
	m.selectedService = 0
	return *m
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
	cmds := make([]tea.Cmd, len(m.inputs)+len(m.newServiceInputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	for i := range m.newServiceInputs {
		m.newServiceInputs[i], cmds[i+len(m.inputs)] = m.newServiceInputs[i].Update(msg)
	}
	return cmds
}

func (m *Model) updateFocus(urlBlockSize int) []tea.Cmd {

	cmds := make([]tea.Cmd, len(m.inputs)+len(m.newServiceInputs))
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

	for i := 0; i < len(m.newServiceInputs); i++ {
		index := len(m.inputs) + i
		if i == m.focusIndex-len(m.inputs)-urlBlockSize-3-len(m.serviceNames) {
			// Set focused state
			cmds[index] = m.newServiceInputs[i].Focus()
			continue
		}
		// Remove focused state
		m.newServiceInputs[i].Blur()
		m.newServiceInputs[i].PromptStyle = noStyle
		m.newServiceInputs[i].TextStyle = noStyle
	}

	return cmds
}

func createKeychain(m *Model) Model {
	url := m.inputs[0].Value()
	accessSeed, err := archethic.MaybeConvertToHex(m.inputs[1].Value())
	if err != nil {
		m.feedback = err.Error()
		return *m
	}
	originPrivateKey, _ := hex.DecodeString("01019280BDB84B8F8AEDBA205FE3552689964A5626EE2C60AA10E3BF22A91A036009")

	publicKey, _, err := archethic.DeriveKeypair(accessSeed, 0, archethic.ED25519)
	if err != nil {
		m.feedback = err.Error()
		return *m
	}

	randomSeed := make([]byte, 32)
	rand.Read(randomSeed)

	keychain := archethic.NewKeychain(randomSeed)
	keychain.AddService("uco", "m/650'/0/0", archethic.ED25519, archethic.SHA256)
	keychain.AddAuthorizedPublicKey(publicKey)

	accessAddress, err := archethic.DeriveAddress(accessSeed, 1, archethic.ED25519, archethic.SHA256)
	if err != nil {
		m.feedback = err.Error()
		return *m
	}
	keychainAddress, err := archethic.DeriveAddress(randomSeed, 1, archethic.ED25519, archethic.SHA256)
	if err != nil {
		m.feedback = err.Error()
		return *m
	}

	keychainTx, err := archethic.NewKeychainTransaction(keychain, 0)
	if err != nil {
		m.feedback = err.Error()
		return *m
	}
	keychainTx.OriginSign(originPrivateKey)

	client := archethic.NewAPIClient(url)
	accessKeychain, _ := archethic.GetKeychain(accessSeed, *client)
	if accessKeychain != nil {
		m.feedback = "Keychain access already exists"
		return *m
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
			m.feedback += "\nKeychain access transaction confirmed."
			ts2.Unsubscribe("confirmation")
			m.keychainAccessTransactionAddress = fmt.Sprintf("%s/explorer/transaction/%x", url, accessAddress)
		})
		ts2.AddOnError(func(senderContext, message string) {
			m.feedback += fmt.Sprintf("\nAccess transaction error: %s", message)
			ts.Unsubscribe("error")
		})
		ts2.SendTransaction(accessTx, 100, 60)
		ts.Unsubscribe("confirmation")
	})
	ts.AddOnError(func(senderContext, message string) {
		m.feedback += fmt.Sprintf("Keychain transaction error: %s", message)
		ts.Unsubscribe("error")
	})
	ts.SendTransaction(keychainTx, 100, 60)
	return *m
}

func addServiceToKeychain(m *Model, accessSeed []byte, client archethic.APIClient, serviceName string, serviceDerivationPath string) {
	updateKeychain(m, accessSeed, client, func(keychain *archethic.Keychain) {
		keychain.AddService(serviceName, serviceDerivationPath, archethic.ED25519, archethic.SHA256)
	})
}

func removeServiceFromKeychain(m *Model, accessSeed []byte, client archethic.APIClient, serviceName string) {
	updateKeychain(m, accessSeed, client, func(keychain *archethic.Keychain) {
		keychain.RemoveService(serviceName)
	})
}

func updateKeychain(m *Model, accessSeed []byte, client archethic.APIClient, updateFunc func(*archethic.Keychain)) error {
	keychain, err := archethic.GetKeychain(accessSeed, client)
	if err != nil {
		return err
	}
	updateFunc(keychain)

	keychainGenesisAddress, err := archethic.DeriveAddress(keychain.Seed, 0, archethic.ED25519, archethic.SHA256)
	if err != nil {
		return err
	}
	addressHex := hex.EncodeToString(keychainGenesisAddress)
	transactionChainIndex := client.GetLastTransactionIndex(addressHex)
	transaction, err := archethic.NewKeychainTransaction(keychain, uint32(transactionChainIndex))
	if err != nil {
		return err
	}
	originPrivateKey, _ := hex.DecodeString("01019280BDB84B8F8AEDBA205FE3552689964A5626EE2C60AA10E3BF22A91A036009")
	transaction.OriginSign(originPrivateKey)

	var returnedError error
	returnedError = nil

	ts := archethic.NewTransactionSender(&client)
	ts.AddOnRequiredConfirmation(func(nbConf int) {
		m.feedback = "\nKeychain's transaction confirmed."
	})
	ts.AddOnError(func(senderContext, message string) {
		returnedError = errors.New(message)
		ts.Unsubscribe("error")
	})
	ts.SendTransaction(transaction, 100, 60)

	return returnedError
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString("> Node endpoint:\n")
	b.WriteString(urlView(m))
	for i := range m.inputs {
		b.WriteRune('\n')
		b.WriteString(m.inputs[i].View())
	}

	if m.showSpinnerCreate {
		b.WriteString("\n\n")
		b.WriteString(m.Spinner.View())
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

	if m.showSpinnerAccess {
		b.WriteString("\n\n")
		b.WriteString(m.Spinner.View())
	}
	button := &accessBlurredButton
	if m.focusIndex == len(m.inputs)+5 {
		button = &accessFocusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	if m.keychain != nil {
		var b2 strings.Builder
		b2.WriteString("--------------------\n")
		b2.WriteString("KEYCHAIN MANAGEMENT:\n")
		b2.WriteString("--------------------\n\n")
		b2.WriteString("List of services:\n")
		for i, k := range m.serviceNames {

			var u string
			if m.selectedService == i {
				u = "(•) "
			} else {
				u = "( ) "
			}
			u += k + " : " + m.keychain.Services[k].DerivationPath + "\n"
			if m.focusIndex == i+len(m.inputs)+6 {
				b2.WriteString(focusedStyle.Render(u))
			} else {
				b2.WriteString(u)
			}
			b2.WriteString("\n")
		}
		b2.WriteString(helpStyle.Render("press 'enter' to select or 'd' to delete "))

		if m.showSpinnerDeleteService {
			b2.WriteString("\n\n")
			b2.WriteString(m.Spinner.View())
		}

		if len(m.serviceNames) > 0 {
			button := &createTransactionaccessBlurredButton
			if m.focusIndex == len(m.inputs)+len(m.serviceNames)+6 {
				button = &createTransactionFocusedButton
			}
			fmt.Fprintf(&b2, "\n\n\n%s\n\n", *button)
		} else {
			b2.WriteString("No service")
		}

		b2.WriteString("Add a service:\n")
		// add fields for service name and derivation path
		for i := range m.newServiceInputs {
			b2.WriteRune('\n')
			b2.WriteString(m.newServiceInputs[i].View())
		}

		if m.showSpinnerCreateService {
			b2.WriteString("\n\n")
			b2.WriteString(m.Spinner.View())
		}

		createServiceButton := &createServiceBlurredButton
		if m.focusIndex == len(m.inputs)+len(m.serviceNames)+6+len(m.newServiceInputs)+1 {
			createServiceButton = &createServiceFocusedButton
		}
		fmt.Fprintf(&b2, "\n\n%s\n\n", *createServiceButton)

		dialogBoxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 10).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

		dialog := lipgloss.Place(150, 9,
			lipgloss.Left, lipgloss.Center,
			dialogBoxStyle.Render(b2.String()),
		)
		b.WriteString(dialog)
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
