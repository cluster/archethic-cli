package keychaincreatetransactionui

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/archethic-foundation/archethic-cli/tui/tuiutils"
	archethic "github.com/archethic-foundation/libgo"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BackMsg bool

type CreateTransactionType int

type CreateTransactionMsg struct {
	ServiceName string
	Seed        string
	Url         string
}

var (
	focusedStyle               = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle               = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle                = focusedStyle.Copy()
	noStyle                    = lipgloss.NewStyle()
	helpStyle                  = blurredStyle.Copy()
	focusedAddAuthKey          = focusedStyle.Copy().Render("[ Add authorization key ]")
	blurredAddAuthKey          = fmt.Sprintf("[ %s ]", blurredStyle.Render("Add authorization key"))
	focusedLoadStorageNouncePK = focusedStyle.Copy().Render("[ Load Storage Nounce Public Key ]")
	blurredLoadStorageNouncePK = fmt.Sprintf("[ %s ]", blurredStyle.Render("Load Storage Nounce Public Key"))
	focusedButton              = focusedStyle.Copy().Render("[ Add ]")
	blurredButton              = fmt.Sprintf("[ %s ]", blurredStyle.Render("Add"))
	urlType                    = []string{"Local", "Testnet", "Mainnet", "Custom"}
	urls                       = map[string]string{
		"Local":   "http://localhost:4000",
		"Testnet": "https://testnet.archethic.net",
		"Mainnet": "https://mainnet.archethic.net",
		"Custom":  ""}
	transactionTypesList = []string{
		"Keychain Access",
		"Keychain",
		"Transfer",
		"Hosting",
		"Token",
		"Data",
		"Contract",
		"Code Proposal",
		"Code Approval",
	}
	transactionTypes = map[string]archethic.TransactionType{
		"Keychain Access": archethic.KeychainAccessType,
		"Keychain":        archethic.KeychainType,
		"Transfer":        archethic.TransferType,
		"Hosting":         archethic.HostingType,
		"Token":           archethic.TokenType,
		"Data":            archethic.DataType,
		"Contract":        archethic.ContractType,
		"Code Proposal":   archethic.CodeProposalType,
		"Code Approval":   archethic.CodeApprovalType,
	}
)

type RenderFunc func(m Model) string
type createTransactionTab int

const (
	MAIN_TAB           createTransactionTab = 0
	UCO_TAB            createTransactionTab = 1
	TOKEN_TAB          createTransactionTab = 2
	RECIPIENTS_TAB     createTransactionTab = 3
	OWNERSHIPS_TAB     createTransactionTab = 4
	CONTENT_TAB        createTransactionTab = 5
	SMART_CONTRACT_TAB createTransactionTab = 6
)

const (
	MAIN_ADD_BUTTON_INDEX         = 17
	FIRST_TRANSACTION_TYPE_INDEX  = 8
	URL_INDEX                     = 4
	TRANSACTION_INDEX_FIELD_INDEX = 7
)

type Model struct {
	Tabs                       []string
	TabContent                 []RenderFunc
	activeTab                  createTransactionTab
	focusInput                 int
	mainInputs                 []textinput.Model
	contentTextAreaInput       textarea.Model
	smartContractTextAreaInput textarea.Model
	ucoInputs                  []textinput.Model
	tokenInputs                []textinput.Model
	recipientsInput            textinput.Model
	ownershipsInputs           []textinput.Model
	transaction                archethic.TransactionBuilder
	secretKey                  []byte
	authorizedKeys             []string
	storageNouncePublicKey     string
	serviceName                string
	serviceMode                bool
	feedback                   string
	selectedUrl                string
}

func New() Model {
	key := make([]byte, 32)
	rand.Read(key)
	m := Model{
		mainInputs:       make([]textinput.Model, 5),
		ucoInputs:        make([]textinput.Model, 2),
		tokenInputs:      make([]textinput.Model, 4),
		ownershipsInputs: make([]textinput.Model, 2),
		focusInput:       0,
		activeTab:        MAIN_TAB,
		transaction:      *archethic.NewTransaction(archethic.KeychainAccessType),
		secretKey:        key,
	}
	tabs := []string{"Main", "UCO Transfers", "Token Transfers", "Recipients", "Ownerships", "Content", "Smart Contract"}
	tabContent := []RenderFunc{main, ucoTransfer, tokenTransfer, recipients, ownerships, content, smartContract}
	m.TabContent = tabContent
	m.Tabs = tabs

	for i := range m.mainInputs {
		t := textinput.New()
		t.CursorStyle = cursorStyle
		switch i {
		case 0:
			t.Prompt = ""
		case 1:
			t.Prompt = "> Access seed\n"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		case 2:
			t.Prompt = "> Elliptic curve\n"
			t.Placeholder = "(default 0)"
			t.CharLimit = 1
			t.Validate = curveValidator
		case 3:
			t.Prompt = "> Index\n"
			t.Placeholder = "(default 0)"
			t.Validate = numberValidator
		case 4:
			t.Prompt = ""
		}
		m.mainInputs[i] = t
	}

	for i := range m.ucoInputs {
		t := textinput.New()
		t.CursorStyle = cursorStyle
		switch i {
		case 0:
			t.Prompt = "> To:\n"
		case 1:
			t.Prompt = "> Amount:\n"
		}

		m.ucoInputs[i] = t
	}

	for i := range m.tokenInputs {
		t := textinput.New()
		t.CursorStyle = cursorStyle
		switch i {
		case 0:
			t.Prompt = "> To:\n"
		case 1:
			t.Prompt = "> Amount:\n"
		case 2:
			t.Prompt = "> Token Address:\n"
		case 3:
			t.Prompt = "> Token ID:\n"
		}

		m.tokenInputs[i] = t
	}

	m.recipientsInput = textinput.New()
	m.recipientsInput.CursorStyle = cursorStyle
	m.recipientsInput.Prompt = "> Recipient address:\n"

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

	m.contentTextAreaInput = textarea.New()
	// passing 0 or a negative number here doesn't seem to work...
	m.contentTextAreaInput.CharLimit = 100000000000
	m.contentTextAreaInput.SetWidth(150)
	m.smartContractTextAreaInput = textarea.New()
	m.smartContractTextAreaInput.CharLimit = 100000000000
	m.smartContractTextAreaInput.SetWidth(150)

	return m
}

func numberValidator(s string) error {
	_, err := strconv.ParseInt(s, 10, 32)
	return err
}

func curveValidator(s string) error {
	val, err := strconv.ParseInt(s, 10, 32)
	if err == nil && (val < 0 || val > 2) {
		return errors.New("number should be >0 and <=2")
	}
	return err
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case CreateTransactionMsg:
		initParams := CreateTransactionMsg(msg)
		m.mainInputs[1].SetValue(initParams.Seed)
		m.mainInputs[0].SetValue(initParams.Url)
		m.serviceName = initParams.ServiceName
		m.serviceMode = m.serviceName != ""
		if m.serviceMode {
			m.focusInput = FIRST_TRANSACTION_TYPE_INDEX
		}
	case tea.KeyMsg:
		event, value := GetState(msg.String(), &m)
		switch event {
		case CONTENT_QUIT_EDIT_MODE:
			m.contentTextAreaInput.Blur()
			return m, nil
		case SMART_CONTRACT_QUIT_EDIT_MODE:
			m.smartContractTextAreaInput.Blur()
			return m, nil
		case BACK_TO_MAIN_MENU:
			return m, func() tea.Msg {
				return BackMsg(true)
			}
		case QUIT:
			return m, tea.Quit
		case SWITCH_NEXT_TAB:
			m.activeTab = createTransactionTab(min(int(m.activeTab)+1, len(m.Tabs)-1))
		case SWITCH_PREVIOUS_TAB:
			m.activeTab = createTransactionTab(max(int(m.activeTab)-1, 0))
		case UPDATE_TRANSACTION_INDEX:
			client := archethic.NewAPIClient(m.mainInputs[0].Value())
			seed := archethic.MaybeConvertToHex(m.mainInputs[1].Value())
			address := archethic.DeriveAddress(seed, 0, getCurve(&m), archethic.SHA256)
			addressHex := hex.EncodeToString(address)
			index := client.GetLastTransactionIndex(addressHex)
			m.mainInputs[3].SetValue(fmt.Sprint(index))
		case UPDATE_URL:
			urlIndex, _ := value.(int)
			u := urlType[urlIndex]
			m.mainInputs[0].SetValue(urls[u])
			m.selectedUrl = u
			m.storageNouncePublicKey = ""
		case UPDATE_TRANSACTION_TYPE:
			typeIndex, _ := value.(string)
			m.transaction.SetType(transactionTypes[typeIndex])
		case SEND_TRANSACTION:
			sendTransaction(&m)
		case ADD_UCO_TRANSFER:
			addUcoTransfer(&m)
		case ADD_TOKEN_TRANSFER:
			addTokenTransfer(&m)
		case ADD_RECIPIENT:
			addRecipient(&m)
		case ADD_AUTHORIZED_KEY:
			addAuthorizedKey(&m)
		case ADD_OWNERSHIP:
			addOwnership(&m)
		case DELETE_UCO_TRANSFER:
			indexToDelete, _ := value.(int)
			deleteUcoTransfer(&m, indexToDelete)
			return m, nil
		case DELETE_TOKEN_TRANSFER:
			indexToDelete, _ := value.(int)
			deleteTokenTransfer(&m, indexToDelete)
			return m, nil
		case DELETE_RECIPIENT:
			indexToDelete, _ := value.(int)
			deleteRecipient(&m, indexToDelete)
			return m, nil
		case DELETE_AUTHORIZED_KEY:
			indexToDelete, _ := value.(int)
			deleteAuthorizedKey(&m, indexToDelete)
			return m, nil
		case DELETE_OWNERSHIP:
			indexToDelete, _ := value.(int)
			deleteOwnership(&m, indexToDelete)
			return m, nil
		case LOAD_STORAGE_NOUNCE_PUBLIC_KEY:
			loadStorageNouncePublicKey(&m)
		case UPDATE_CONTENT:
			m.transaction.SetContent([]byte(m.contentTextAreaInput.Value()))
		case UPDATE_SMART_CONTRACT:
			m.transaction.SetCode(m.smartContractTextAreaInput.Value())
		}

	}
	m, cmds := updateFocus(m)

	cmds = append(cmds, m.updateInputs(msg)...)

	return m, tea.Batch(cmds...)
}

func (m *Model) updateInputs(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd
	switch m.activeTab {
	case MAIN_TAB:
		cmds := make([]tea.Cmd, len(m.mainInputs))
		for i := range m.mainInputs {
			m.mainInputs[i], cmds[i] = m.mainInputs[i].Update(msg)
		}
	case UCO_TAB:
		cmds := make([]tea.Cmd, len(m.ucoInputs))
		for i := range m.ucoInputs {
			m.ucoInputs[i], cmds[i] = m.ucoInputs[i].Update(msg)
		}
	case TOKEN_TAB:
		cmds := make([]tea.Cmd, len(m.tokenInputs))
		for i := range m.tokenInputs {
			m.tokenInputs[i], cmds[i] = m.tokenInputs[i].Update(msg)
		}
	case RECIPIENTS_TAB:
		cmds := make([]tea.Cmd, 1)
		m.recipientsInput, cmds[0] = m.recipientsInput.Update(msg)
	case OWNERSHIPS_TAB:
		cmds := make([]tea.Cmd, len(m.ownershipsInputs))
		for i := range m.ownershipsInputs {
			m.ownershipsInputs[i], cmds[i] = m.ownershipsInputs[i].Update(msg)
		}
	case CONTENT_TAB:
		cmds := make([]tea.Cmd, 1)
		m.contentTextAreaInput, cmds[0] = m.contentTextAreaInput.Update(msg)
	case SMART_CONTRACT_TAB:
		cmds := make([]tea.Cmd, 1)
		m.smartContractTextAreaInput, cmds[0] = m.smartContractTextAreaInput.Update(msg)
	}

	return cmds
}

func updateFocus(m Model) (Model, []tea.Cmd) {
	var cmds []tea.Cmd
	switch m.activeTab {
	case MAIN_TAB:
		cmds := make([]tea.Cmd, len(m.mainInputs))
		for i := 0; i <= len(m.mainInputs)-1; i++ {
			// the first 4 inputs are not focusable fields (node endpoints for URL)
			if i == m.focusInput-len(urlType) {
				// Set focused state
				cmds[i] = m.mainInputs[i].Focus()
				continue
			}
			// Remove focused state
			m.mainInputs[i].Blur()
			m.mainInputs[i].PromptStyle = noStyle
			m.mainInputs[i].TextStyle = noStyle
		}

	case UCO_TAB:

		cmds := make([]tea.Cmd, len(m.ucoInputs))
		for i := 0; i <= len(m.ucoInputs)-1; i++ {
			if i == m.focusInput {
				// Set focused state
				cmds[i] = m.ucoInputs[i].Focus()
				continue
			}
			// Remove focused state
			m.ucoInputs[i].Blur()
			m.ucoInputs[i].PromptStyle = noStyle
			m.ucoInputs[i].TextStyle = noStyle
		}

	case TOKEN_TAB:

		cmds := make([]tea.Cmd, len(m.tokenInputs))
		for i := 0; i <= len(m.tokenInputs)-1; i++ {
			if i == m.focusInput {
				// Set focused state
				cmds[i] = m.tokenInputs[i].Focus()
				continue
			}
			// Remove focused state
			m.tokenInputs[i].Blur()
			m.tokenInputs[i].PromptStyle = noStyle
			m.tokenInputs[i].TextStyle = noStyle
		}

	case RECIPIENTS_TAB:
		if m.focusInput == 0 {
			cmds = append(cmds, m.recipientsInput.Focus())
			m.recipientsInput.PromptStyle = focusedStyle
			m.recipientsInput.TextStyle = focusedStyle
		} else {
			m.recipientsInput.Blur()
			m.recipientsInput.PromptStyle = noStyle
			m.recipientsInput.TextStyle = noStyle
		}
	case OWNERSHIPS_TAB:
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
	case CONTENT_TAB:
		cmds := make([]tea.Cmd, 1)
		cmds[0] = m.contentTextAreaInput.Focus()
	case SMART_CONTRACT_TAB:
		cmds := make([]tea.Cmd, 1)
		cmds[0] = m.smartContractTextAreaInput.Focus()
	}
	return m, cmds
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 5)
	activeTabStyle    = inactiveTabStyle.Copy().Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Left).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

func (m Model) View() string {
	doc := strings.Builder{}

	var renderedTabs []string

	for i, t := range m.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.Tabs)-1, i == int(m.activeTab)
		if isActive {
			style = activeTabStyle.Copy()
		} else {
			style = inactiveTabStyle.Copy()
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).Render(m.TabContent[m.activeTab](m)))
	return docStyle.Render(doc.String())
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getCurve(m *Model) archethic.Curve {
	curveInt, err := strconv.Atoi(m.mainInputs[2].Value())
	if err != nil {
		curveInt = 0
	}
	return archethic.Curve(curveInt)
}

func addUcoTransfer(m *Model) {
	toHex := m.ucoInputs[0].Value()
	to, err := hex.DecodeString(toHex)
	if err != nil {
		panic(err)
	}
	amountStr := m.ucoInputs[1].Value()
	amount, err := strconv.ParseUint(amountStr, 10, 64)
	if err != nil {
		panic(err)
	}
	m.transaction.AddUcoTransfer(to, amount)
	m.ucoInputs[0].SetValue("")
	m.ucoInputs[1].SetValue("")
}

func deleteUcoTransfer(m *Model, indexToDelete int) {
	m.transaction.Data.Ledger.Uco.Transfers = append(m.transaction.Data.Ledger.Uco.Transfers[:indexToDelete], m.transaction.Data.Ledger.Uco.Transfers[indexToDelete+1:]...)
}

func addTokenTransfer(m *Model) {
	toHex := m.tokenInputs[0].Value()
	to, err := hex.DecodeString(toHex)
	if err != nil {
		panic(err)
	}
	amountStr := m.tokenInputs[1].Value()
	amount, err := strconv.ParseUint(amountStr, 10, 64)
	if err != nil {
		panic(err)
	}
	tokenAddressHex := m.tokenInputs[2].Value()
	tokenAddress, err := hex.DecodeString(tokenAddressHex)
	if err != nil {
		panic(err)
	}
	tokenIdStr := m.tokenInputs[3].Value()

	tokenId, err := strconv.Atoi(tokenIdStr)
	if err != nil {
		panic(err)
	}
	m.transaction.AddTokenTransfer(to, tokenAddress, amount, tokenId)
	m.tokenInputs[0].SetValue("")
	m.tokenInputs[1].SetValue("")
	m.tokenInputs[2].SetValue("")
	m.tokenInputs[3].SetValue("")
}

func deleteTokenTransfer(m *Model, indexToDelete int) {
	m.transaction.Data.Ledger.Token.Transfers = append(m.transaction.Data.Ledger.Token.Transfers[:indexToDelete], m.transaction.Data.Ledger.Token.Transfers[indexToDelete+1:]...)
}

func addRecipient(m *Model) {
	recipientHex := m.recipientsInput.Value()
	recipient, err := hex.DecodeString(recipientHex)
	if err != nil {
		panic(err)
	}
	m.transaction.AddRecipient(recipient)
	m.recipientsInput.SetValue("")
}

func deleteRecipient(m *Model, indexToDelete int) {
	m.transaction.Data.Recipients = append(m.transaction.Data.Recipients[:indexToDelete], m.transaction.Data.Recipients[indexToDelete+1:]...)
}

func addOwnership(m *Model) {

	secret := archethic.MaybeConvertToHex(m.ownershipsInputs[0].Value())

	if m.ownershipsInputs[1].Value() != "" {
		addAuthorizedKey(m)
	}

	cipher := archethic.AesEncrypt([]byte(secret), m.secretKey)
	authorizedKeys := make([]archethic.AuthorizedKey, len(m.authorizedKeys))
	for i, key := range m.authorizedKeys {
		keyByte, err := hex.DecodeString(key)
		if err != nil {
			panic(err)
		}
		authorizedKeys[i] = archethic.AuthorizedKey{
			PublicKey:          keyByte,
			EncryptedSecretKey: archethic.EcEncrypt(m.secretKey, keyByte),
		}
	}
	m.transaction.AddOwnership(cipher, authorizedKeys)
	m.authorizedKeys = []string{}
	m.ownershipsInputs[0].SetValue("")
	m.ownershipsInputs[1].SetValue("")
}

func deleteOwnership(m *Model, indexToDelete int) {
	m.transaction.Data.Ownerships = append(m.transaction.Data.Ownerships[:indexToDelete], m.transaction.Data.Ownerships[indexToDelete+1:]...)
}

func addAuthorizedKey(m *Model) {
	authorizedKey := m.ownershipsInputs[1].Value()
	m.authorizedKeys = append(m.authorizedKeys, authorizedKey)
	m.ownershipsInputs[1].SetValue("")
}

func deleteAuthorizedKey(m *Model, indexToDelete int) {
	m.authorizedKeys = append(m.authorizedKeys[:indexToDelete], m.authorizedKeys[indexToDelete+1:]...)
}

func loadStorageNouncePublicKey(m *Model) {

	if m.storageNouncePublicKey == "" {
		client := archethic.NewAPIClient(m.mainInputs[0].Value())
		m.storageNouncePublicKey = client.GetStorageNoncePublicKey()
	}

	m.ownershipsInputs[1].SetValue(m.storageNouncePublicKey)
}

func buildKeychainTransaction(seed []byte, client *archethic.APIClient, m *Model) {
	keychain := archethic.GetKeychain(seed, *client)

	m.transaction.Version = uint32(keychain.Version)

	genesisAddress := keychain.DeriveAddress(m.serviceName, 0)

	index := client.GetLastTransactionIndex(hex.EncodeToString(genesisAddress))

	m.transaction = keychain.BuildTransaction(m.transaction, m.serviceName, uint8(index))
}

func sendTransaction(m *Model) {
	m.feedback = ""
	if len(m.transaction.Data.Code) > 0 {

		ownershipIndex := -1
		for i, ownership := range m.transaction.Data.Ownerships {
			decodedSecret := string(archethic.AesDecrypt(ownership.Secret, m.secretKey))

			if reflect.DeepEqual(decodedSecret, string(archethic.MaybeConvertToHex(m.mainInputs[1].Value()))) {
				ownershipIndex = i
				break
			}
		}

		if ownershipIndex == -1 {
			m.feedback = "You need to create an ownership with the transaction seed as secret and authorize node public key to let nodes generate new transaction from your smart contract"
		} else {
			authorizedKeyIndex := -1
			for i, authKey := range m.transaction.Data.Ownerships[ownershipIndex].AuthorizedKeys {
				if strings.ToUpper(hex.EncodeToString(authKey.PublicKey)) == m.storageNouncePublicKey {
					authorizedKeyIndex = i
					break
				}
			}

			if authorizedKeyIndex == -1 {
				m.feedback = "You need to create an ownership with the transaction seed as secret and authorize node public key to let nodes generate new transaction from your smart contract"
			}
		}
	}

	if m.feedback == "" {
		url := m.mainInputs[0].Value()
		client := archethic.NewAPIClient(url)

		seed := archethic.MaybeConvertToHex(m.mainInputs[1].Value())

		if m.serviceMode {
			buildKeychainTransaction(seed, client, m)
		} else {
			index, err := strconv.ParseUint(m.mainInputs[3].Value(), 10, 32)
			if err != nil {
				index = 0
			}
			curve := getCurve(m)
			m.transaction.Build(seed, uint32(index), curve, archethic.SHA256)

		}
		originPrivateKey, _ := hex.DecodeString("01019280BDB84B8F8AEDBA205FE3552689964A5626EE2C60AA10E3BF22A91A036009")
		m.transaction.OriginSign(originPrivateKey)

		ts := archethic.NewTransactionSender(client)
		ts.AddOnSent(func() {
			m.feedback = "Transaction sent: " + url + "/explorer/transaction/" + strings.ToUpper(hex.EncodeToString(m.transaction.Address))
		})

		ts.AddOnError(func(sender, message string) {
			m.feedback = "Transaction error: " + message
		})

		ts.SendTransaction(&m.transaction, 100, 60)
	}
}

// Functions to build the view

// View of the main tab
func main(m Model) string {
	var b strings.Builder
	// Only display the node endpoint, seed, curve and index fields if we are not building a transaction for a service
	if m.serviceMode {
		b.WriteString("Creating transaction for service " + m.serviceName + "\n\n")
	} else {
		b.WriteString("> Node endpoint:\n")
		b.WriteString(urlView(m))
		// url field
		b.WriteString(m.mainInputs[0].View() + "\n\n")
		// seed field
		b.WriteString(m.mainInputs[1].View() + "\n\n")
		// curve field
		b.WriteString(m.mainInputs[2].View() + "\n\n")
		for j := 0; j <= 2; j++ {
			b.WriteString("\t (" + strconv.Itoa(j) + ") " + tuiutils.GetCurveName(archethic.Curve(j)) + "\n")
		}
		// index field
		b.WriteString(m.mainInputs[3].View() + "\n\n")
	}

	// transaction type field
	b.WriteString("> Transaction type:\n")
	b.WriteString(transactionTypeView(m))

	// send transaction button
	button := &blurredButton
	if m.focusInput == MAIN_ADD_BUTTON_INDEX {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	b.WriteString(m.feedback)
	return b.String()
}

// URL part of the main tab
// looping through the urlType array to display the different options
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
		if i == m.focusInput {
			s.WriteString(focusedStyle.Render(u))
		} else {
			s.WriteString(u)
		}
		s.WriteString("\n")
	}

	return s.String()
}

// Transaction type part of the main tab
func transactionTypeView(m Model) string {
	s := strings.Builder{}

	for i, t := range transactionTypesList {
		var u string
		if m.transaction.TxType == transactionTypes[t] {
			u = "(•) "
		} else {
			u = "( ) "
		}
		u += t
		if m.focusInput == i+FIRST_TRANSACTION_TYPE_INDEX {
			s.WriteString(focusedStyle.Render(u))
		} else {
			s.WriteString(u)
		}
		s.WriteString("\n")
	}

	return s.String()
}

// View of the UCO transfer tab
func ucoTransfer(m Model) string {
	var b strings.Builder
	for i := range m.ucoInputs {
		b.WriteString(m.ucoInputs[i].View())
		if i < len(m.ucoInputs)-1 {
			b.WriteRune('\n')
		}
	}
	button := &blurredButton
	if m.focusInput == len(m.ucoInputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	startCount := len(m.ucoInputs) + 1 // +1 for the button
	for i, t := range m.transaction.Data.Ledger.Uco.Transfers {
		transfer := fmt.Sprintf("%s: %d\n", hex.EncodeToString(t.To), t.Amount)
		if m.focusInput == startCount+i {
			b.WriteString(focusedStyle.Render(transfer))
			continue
		} else {
			b.WriteString(transfer)
		}
	}
	if len(m.transaction.Data.Ledger.Uco.Transfers) > 0 {
		b.WriteString(helpStyle.Render("\npress 'd' to delete the selected UCO transfer "))
	}
	return b.String()
}

// View of the token transfer tab
func tokenTransfer(m Model) string {
	var b strings.Builder
	for i := range m.tokenInputs {
		b.WriteString(m.tokenInputs[i].View())
		if i < len(m.tokenInputs)-1 {
			b.WriteRune('\n')
		}
	}
	button := &blurredButton
	if m.focusInput == len(m.tokenInputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	startCount := len(m.tokenInputs) + 1 // +1 for the button
	for i, t := range m.transaction.Data.Ledger.Token.Transfers {
		transfer := fmt.Sprintf("%s : %d - %s %d \n", hex.EncodeToString(t.To), t.Amount, hex.EncodeToString(t.TokenAddress), t.TokenId)
		if m.focusInput == startCount+i {
			b.WriteString(focusedStyle.Render(transfer))
			continue
		} else {
			b.WriteString(transfer)
		}
	}
	if len(m.transaction.Data.Ledger.Token.Transfers) > 0 {
		b.WriteString(helpStyle.Render("\npress 'd' to delete the selected token transfer "))
	}
	return b.String()
}

// View of the recipients tab
func recipients(m Model) string {
	var b strings.Builder
	b.WriteString(m.recipientsInput.View())
	b.WriteRune('\n')

	button := &blurredButton
	if m.focusInput == 1 {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	startCount := 2 // 1 for the input, 1 for the button
	for i, t := range m.transaction.Data.Recipients {
		recipient := fmt.Sprintf("%s\n", hex.EncodeToString(t))
		if m.focusInput == startCount+i {
			b.WriteString(focusedStyle.Render(recipient))
			continue
		} else {
			b.WriteString(recipient)
		}
	}
	if len(m.transaction.Data.Recipients) > 0 {
		b.WriteString(helpStyle.Render("\npress 'd' to delete the selected recipient "))
	}
	return b.String()
}

// View of the ownerships tab
func ownerships(m Model) string {
	var b strings.Builder
	for i := range m.ownershipsInputs {
		b.WriteString(m.ownershipsInputs[i].View())
		if i < len(m.ownershipsInputs)-1 {
			b.WriteRune('\n')
		}
	}

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
	fmt.Fprintf(&b, "\n\n%s", *buttonLoadStorageNouncePK)

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

// View of the content tab
func content(m Model) string {
	var b strings.Builder

	b.WriteString(m.contentTextAreaInput.View())

	if m.contentTextAreaInput.Focused() {
		b.WriteString(helpStyle.Render("\npress 'esc' to exit edit mode "))
	}

	return b.String()
}

// View of the smart contract tab
func smartContract(m Model) string {
	var b strings.Builder

	b.WriteString(m.smartContractTextAreaInput.View())
	if m.smartContractTextAreaInput.Focused() {
		b.WriteString(helpStyle.Render("\npress 'esc' to exit edit mode "))
	}
	return b.String()
}
