package keychaincreatetransactionui

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	archethic "github.com/archethic-foundation/libgo"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type OwnershipsModel struct {
	focusInput             int
	ownershipsInputs       []textinput.Model
	authorizedKeys         []string
	url                    string
	storageNouncePublicKey string
	secretKey              []byte
	transaction            *archethic.TransactionBuilder
	feedback               string
	showSpinner            bool
	Spinner                spinner.Model
	IsInit                 bool
}

type AddOwnership struct {
	Cipher         []byte
	AuthorizedKeys []archethic.AuthorizedKey
	cmds           []tea.Cmd
}

type SendUpdateStorageNouncePublicKey struct {
	OwnershipsModel OwnershipsModel
}

type UpdateStorageNouncePublicKey struct {
	StorageNouncePublicKey string
}
type DeleteOwnership struct {
	IndexToDelete int
}

func NewOwnershipsModel(secretKey []byte, transaction *archethic.TransactionBuilder) OwnershipsModel {

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m := OwnershipsModel{
		ownershipsInputs: make([]textinput.Model, 2),
		secretKey:        secretKey,
		transaction:      transaction,
		Spinner:          s,
	}
	for i := range m.ownershipsInputs {
		t := textinput.New()
		t.CursorStyle = cursorStyle
		switch i {
		case 0:
			t.Prompt = "> Secret:\n"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		case 1:
			t.Prompt = "> Authorization key:\n"
		}

		m.ownershipsInputs[i] = t
	}
	return m
}

func (m OwnershipsModel) Init() tea.Cmd {
	return m.Spinner.Tick
}

func (m OwnershipsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {

		case "up", "down":
			updateOwnershipsFocusInput(&m, keypress)

		case "enter":

			switch m.focusInput {

			// add authorized key
			case len(m.ownershipsInputs) + len(m.authorizedKeys):
				err := addAuthorizedKey(&m)
				if err != nil {
					m.feedback = fmt.Sprintf("%s", err)
					return m, nil
				}
				// load storage nounce public key
			case len(m.ownershipsInputs) + len(m.authorizedKeys) + 1:
				if m.url == "" {
					m.feedback = "Please select a node endpoint in the main tab"
					return m, nil
				}
				if m.storageNouncePublicKey == "" {
					m.showSpinner = true
					return m, func() tea.Msg {
						return loadStorageNouncePublicKey(m)
					}
				}
				m.ownershipsInputs[1].SetValue(m.storageNouncePublicKey)
				return m, func() tea.Msg {
					return UpdateStorageNouncePublicKey{StorageNouncePublicKey: m.storageNouncePublicKey}
				}
				//add ownership
			case len(m.ownershipsInputs) + len(m.authorizedKeys) + 2:

				secret, err := archethic.MaybeConvertToHex(m.ownershipsInputs[0].Value())
				if err != nil {
					m.feedback = fmt.Sprintf("%s", err)
					return m, nil
				}

				if m.ownershipsInputs[1].Value() != "" {
					err := addAuthorizedKey(&m)
					if err != nil {
						m.feedback = fmt.Sprintf("%s", err)
						return m, nil
					}
				}

				cipher, err := archethic.AesEncrypt([]byte(secret), m.secretKey)
				if err != nil {
					m.feedback = fmt.Sprintf("%s", err)
					return m, nil
				}
				authorizedKeys := make([]archethic.AuthorizedKey, len(m.authorizedKeys))
				for i, key := range m.authorizedKeys {
					keyByte, err := hex.DecodeString(key)
					if err != nil {
						m.feedback = fmt.Sprintf("%s", err)
						return m, nil
					}
					encrypedSecretKey, err := archethic.EcEncrypt(m.secretKey, keyByte)
					if err != nil {
						m.feedback = fmt.Sprintf("%s", err)
						return m, nil
					}
					authorizedKeys[i] = archethic.AuthorizedKey{
						PublicKey:          keyByte,
						EncryptedSecretKey: encrypedSecretKey,
					}
				}

				m.authorizedKeys = []string{}
				m.ownershipsInputs[0].SetValue("")
				m.ownershipsInputs[1].SetValue("")
				m, cmds := updateOwnershipsFocus(m)
				cmds = append(cmds, m.updateOwnershipsInputs(msg)...)
				return m, func() tea.Msg {
					return AddOwnership{Cipher: cipher, AuthorizedKeys: authorizedKeys, cmds: cmds}
				}
			}

		case "d":

			if m.focusInput > len(m.ownershipsInputs)-1 && m.focusInput < len(m.ownershipsInputs)+len(m.authorizedKeys) {
				indexToDelete := m.focusInput - len(m.ownershipsInputs)
				m.focusInput--
				m.authorizedKeys = append(m.authorizedKeys[:indexToDelete], m.authorizedKeys[indexToDelete+1:]...)
				return m, nil
			} else if m.focusInput > len(m.ownershipsInputs)+len(m.authorizedKeys)+2 {
				indexToDelete := m.focusInput - len(m.ownershipsInputs) - len(m.authorizedKeys) - 3
				m.focusInput--
				return m, func() tea.Msg {
					return DeleteOwnership{indexToDelete}
				}
			}

		}
	case UpdateStorageNouncePublicKey:
		m.showSpinner = false
		return m, nil
	default:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}
	m, cmds := updateOwnershipsFocus(m)
	cmds = append(cmds, m.updateOwnershipsInputs(msg)...)

	return m, tea.Batch(cmds...)
}

func loadStorageNouncePublicKey(m OwnershipsModel) UpdateStorageNouncePublicKey {
	client := archethic.NewAPIClient(m.url)
	var err error
	m.storageNouncePublicKey, err = client.GetStorageNoncePublicKey()
	if err != nil {
		m.feedback = fmt.Sprintf("%s", err)
	}
	m.ownershipsInputs[1].SetValue(m.storageNouncePublicKey)
	m.showSpinner = false
	return UpdateStorageNouncePublicKey{StorageNouncePublicKey: m.storageNouncePublicKey}
}

func addAuthorizedKey(m *OwnershipsModel) error {
	authorizedKey := m.ownershipsInputs[1].Value()
	_, err := hex.DecodeString(authorizedKey)
	if err != nil {
		return errors.New("invalid authorization key")
	}
	m.authorizedKeys = append(m.authorizedKeys, authorizedKey)
	m.ownershipsInputs[1].SetValue("")
	return nil
}

func (m *OwnershipsModel) updateOwnershipsInputs(msg tea.Msg) []tea.Cmd {

	cmds := make([]tea.Cmd, len(m.ownershipsInputs))
	for i := range m.ownershipsInputs {
		m.ownershipsInputs[i], cmds[i] = m.ownershipsInputs[i].Update(msg)
	}

	return cmds
}

func updateOwnershipsFocus(m OwnershipsModel) (OwnershipsModel, []tea.Cmd) {

	cmds := make([]tea.Cmd, len(m.ownershipsInputs))
	for i := 0; i <= len(m.ownershipsInputs)-1; i++ {
		if i == m.focusInput {
			// Set focused state
			cmds[i] = m.ownershipsInputs[i].Focus()
			continue
		}
		// Remove focused state
		m.ownershipsInputs[i].Blur()
		m.ownershipsInputs[i].PromptStyle = noStyle
		m.ownershipsInputs[i].TextStyle = noStyle
	}
	return m, cmds
}

func updateOwnershipsFocusInput(m *OwnershipsModel, keypress string) {
	if keypress == "up" {
		m.focusInput--
	} else {
		m.focusInput++
	}

	if m.focusInput > len(m.ownershipsInputs)+len(m.authorizedKeys)+len(m.transaction.Data.Ownerships)+2 { // 2 for the 3 buttons
		m.focusInput = 0
	} else if m.focusInput < 0 {
		m.focusInput = len(m.ownershipsInputs) + len(m.authorizedKeys) + len(m.transaction.Data.Ownerships) + 2
	}

}

func (m *OwnershipsModel) SwitchTab() (OwnershipsModel, []tea.Cmd) {
	m.focusInput = 0
	m2, cmds := updateOwnershipsFocus(*m)
	return m2, cmds
}

func (m OwnershipsModel) View() string {
	var b strings.Builder
	for i := range m.ownershipsInputs {
		b.WriteString(m.ownershipsInputs[i].View())
		if i < len(m.ownershipsInputs)-1 {
			b.WriteRune('\n')
		}
	}
	b.WriteRune('\n')
	b.WriteString(m.feedback)
	if len(m.authorizedKeys) > 0 {
		b.WriteString("\nList of authorized keys to add:\n")
		for i := range m.authorizedKeys {
			if m.focusInput == len(m.ownershipsInputs)+i {
				b.WriteString(focusedStyle.Render(m.authorizedKeys[i]))
			} else {
				b.WriteString(m.authorizedKeys[i])
			}
			b.WriteRune('\n')
		}
		b.WriteString(helpStyle.Render("\npress 'd' to delete the selected authorized key "))
	}

	buttonAddAuthKey := &blurredAddAuthKey
	if m.focusInput == len(m.ownershipsInputs)+len(m.authorizedKeys) {
		buttonAddAuthKey = &focusedAddAuthKey
	}
	fmt.Fprintf(&b, "\n\n%s", *buttonAddAuthKey)

	buttonLoadStorageNouncePK := &blurredLoadStorageNouncePK
	if m.focusInput == len(m.ownershipsInputs)+len(m.authorizedKeys)+1 {
		buttonLoadStorageNouncePK = &focusedLoadStorageNouncePK
	}
	b.WriteString("\n\n")
	if m.showSpinner {
		b.WriteString(m.Spinner.View())
	}
	fmt.Fprintf(&b, "%s", *buttonLoadStorageNouncePK)

	button := &blurredButton
	if m.focusInput == len(m.ownershipsInputs)+len(m.authorizedKeys)+2 {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	startCount := len(m.ownershipsInputs) + len(m.authorizedKeys) + 3 // +3 for the buttons
	for i, o := range m.transaction.Data.Ownerships {
		ownerships := "**** "
		for j := range o.AuthorizedKeys {
			keyHex := hex.EncodeToString(o.AuthorizedKeys[j].PublicKey)
			ownerships += fmt.Sprintf("%s\n", keyHex)
		}
		if m.focusInput == startCount+i {
			b.WriteString(focusedStyle.Render(ownerships))
			continue
		} else {
			b.WriteString(ownerships)
		}
	}
	if len(m.transaction.Data.Ownerships) > 0 {
		b.WriteString(helpStyle.Render("\npress 'd' to delete the selected ownership "))
	}
	return b.String()
}

func (m *OwnershipsModel) SetUrl(url string) {
	m.url = url
	m.storageNouncePublicKey = ""
}
