package keychaincreatetransactionui

type createTransactionState int

const (
	CONTENT_QUIT_EDIT_MODE createTransactionState = iota
	SMART_CONTRACT_QUIT_EDIT_MODE
	BACK_TO_MAIN_MENU
	QUIT
	SWITCH_NEXT_TAB
	SWITCH_PREVIOUS_TAB
	UPDATE_TRANSACTION_INDEX
	UPDATE_URL
	UPDATE_TRANSACTION_TYPE
	SEND_TRANSACTION
	ADD_UCO_TRANSFER
	ADD_TOKEN_TRANSFER
	ADD_RECIPIENT
	ADD_AUTHORIZED_KEY
	ADD_OWNERSHIP
	DELETE_UCO_TRANSFER
	DELETE_TOKEN_TRANSFER
	DELETE_RECIPIENT
	DELETE_AUTHORIZED_KEY
	DELETE_OWNERSHIP
	LOAD_STORAGE_NOUNCE_PUBLIC_KEY
	UPDATE_CONTENT
	UPDATE_SMART_CONTRACT
	CONTINUE
)

func GetState(msg string, m *Model) (createTransactionState, interface{}) {

	switch keypress := msg; keypress {
	case "esc":
		if m.activeTab == CONTENT_TAB && m.contentTextAreaInput.Focused() {
			return CONTENT_QUIT_EDIT_MODE, nil
		}

		if m.activeTab == SMART_CONTRACT_TAB && m.smartContractTextAreaInput.Focused() {
			return SMART_CONTRACT_QUIT_EDIT_MODE, nil
		}
		return BACK_TO_MAIN_MENU, nil

	case "ctrl+c":
		return QUIT, nil
	case "right", "tab":
		// switch to the next tab except if the user is editing the content or the smart contract
		if (m.activeTab == CONTENT_TAB && !m.contentTextAreaInput.Focused()) ||
			(m.activeTab == SMART_CONTRACT_TAB && !m.smartContractTextAreaInput.Focused()) ||
			(m.activeTab != CONTENT_TAB && m.activeTab != SMART_CONTRACT_TAB) {
			m.focusInput = 0
			return SWITCH_NEXT_TAB, nil
		}

	case "left", "shift+tab":
		// switch to the previous tab except if the user is editing the content or the smart contract
		if (m.activeTab == CONTENT_TAB && !m.contentTextAreaInput.Focused()) ||
			(m.activeTab == SMART_CONTRACT_TAB && !m.smartContractTextAreaInput.Focused()) ||
			(m.activeTab != CONTENT_TAB && m.activeTab != SMART_CONTRACT_TAB) {
			m.focusInput = 0
			return SWITCH_PREVIOUS_TAB, nil
		}

	case "up", "down":
		updateFocusInput(m, keypress)
		// if the seed or the curve are blured, they are probably updated, so update the transaction index
		if m.activeTab == MAIN_TAB && (m.focusInput == URL_INDEX || m.focusInput == TRANSACTION_INDEX_FIELD_INDEX) {
			return UPDATE_TRANSACTION_INDEX, nil
		}

	case "enter":
		switch m.activeTab {
		case MAIN_TAB:

			if m.focusInput < URL_INDEX {
				value := m.focusInput
				m.focusInput = URL_INDEX
				return UPDATE_URL, value
			} else if m.focusInput > TRANSACTION_INDEX_FIELD_INDEX && m.focusInput < MAIN_ADD_BUTTON_INDEX {
				newType := transactionTypesList[m.focusInput-FIRST_TRANSACTION_TYPE_INDEX]
				m.focusInput = MAIN_ADD_BUTTON_INDEX
				return UPDATE_TRANSACTION_TYPE, newType
			} else if m.focusInput == MAIN_ADD_BUTTON_INDEX {
				return SEND_TRANSACTION, nil
			}

		case UCO_TAB:
			if m.focusInput == len(m.ucoInputs) {
				return ADD_UCO_TRANSFER, nil
			}
		case TOKEN_TAB:
			if m.focusInput == len(m.tokenInputs) {
				return ADD_TOKEN_TRANSFER, nil
			}
		case RECIPIENTS_TAB:
			if m.focusInput == 1 || m.focusInput == 0 {
				return ADD_RECIPIENT, nil
			}
		case OWNERSHIPS_TAB:
			switch m.focusInput {
			case len(m.ownershipsInputs) + len(m.authorizedKeys):
				return ADD_AUTHORIZED_KEY, nil
			case len(m.ownershipsInputs) + len(m.authorizedKeys) + 1:
				return LOAD_STORAGE_NOUNCE_PUBLIC_KEY, nil
			case len(m.ownershipsInputs) + len(m.authorizedKeys) + 2:
				return ADD_OWNERSHIP, nil
			}
		}
	case "d":
		switch m.activeTab {
		case UCO_TAB:
			if m.focusInput > len(m.ucoInputs) {
				indexToDelete := m.focusInput - len(m.ucoInputs) - 1
				m.focusInput--
				return DELETE_UCO_TRANSFER, indexToDelete
			}
		case TOKEN_TAB:
			if m.focusInput > len(m.tokenInputs) {
				indexToDelete := m.focusInput - len(m.tokenInputs) - 1
				m.focusInput--
				return DELETE_TOKEN_TRANSFER, indexToDelete
			}
		case RECIPIENTS_TAB:
			if m.focusInput > 1 {
				indexToDelete := m.focusInput - 2
				m.focusInput--
				return DELETE_RECIPIENT, indexToDelete
			}
		case OWNERSHIPS_TAB:
			if m.focusInput > len(m.ownershipsInputs)-1 && m.focusInput < len(m.ownershipsInputs)+len(m.authorizedKeys) {
				indexToDelete := m.focusInput - len(m.ownershipsInputs)
				m.focusInput--
				return DELETE_AUTHORIZED_KEY, indexToDelete
			} else if m.focusInput > len(m.ownershipsInputs)+len(m.authorizedKeys)+2 {
				indexToDelete := m.focusInput - len(m.ownershipsInputs) - len(m.authorizedKeys) - 3
				m.focusInput--
				return DELETE_OWNERSHIP, indexToDelete
			}
		}
	default:
		switch m.activeTab {
		case CONTENT_TAB:
			return UPDATE_CONTENT, nil
		case SMART_CONTRACT_TAB:
			return UPDATE_SMART_CONTRACT, nil
		}
	}
	return CONTINUE, nil
}

func updateFocusInput(m *Model, keypress string) {
	if keypress == "up" {
		m.focusInput--
	} else {
		m.focusInput++
	}
	switch m.activeTab {
	case MAIN_TAB:
		if m.serviceMode {
			if m.focusInput > MAIN_ADD_BUTTON_INDEX {
				m.focusInput = FIRST_TRANSACTION_TYPE_INDEX
			} else if m.focusInput < FIRST_TRANSACTION_TYPE_INDEX {
				m.focusInput = MAIN_ADD_BUTTON_INDEX
			}
		} else {
			if m.focusInput > MAIN_ADD_BUTTON_INDEX {
				m.focusInput = 0
			} else if m.focusInput < 0 {
				m.focusInput = MAIN_ADD_BUTTON_INDEX
			}
		}

	case UCO_TAB:
		if m.focusInput > len(m.ucoInputs)+len(m.transaction.Data.Ledger.Uco.Transfers) {
			m.focusInput = 0
		} else if m.focusInput < 0 {
			m.focusInput = len(m.ucoInputs) + len(m.transaction.Data.Ledger.Uco.Transfers)
		}
	case TOKEN_TAB:
		if m.focusInput > len(m.tokenInputs)+len(m.transaction.Data.Ledger.Token.Transfers) {
			m.focusInput = 0
		} else if m.focusInput < 0 {
			m.focusInput = len(m.tokenInputs) + len(m.transaction.Data.Ledger.Token.Transfers)
		}
	case RECIPIENTS_TAB:
		// 1 because : first input [0] is the recipient address and second [1] is the button
		if m.focusInput > 1+len(m.transaction.Data.Recipients) {
			m.focusInput = 0
		} else if m.focusInput < 0 {
			m.focusInput = 1 + len(m.transaction.Data.Recipients)
		}
	case OWNERSHIPS_TAB:
		if m.focusInput > len(m.ownershipsInputs)+len(m.authorizedKeys)+len(m.transaction.Data.Ownerships)+2 { // 2 for the 3 buttons
			m.focusInput = 0
		} else if m.focusInput < 0 {
			m.focusInput = len(m.ownershipsInputs) + len(m.authorizedKeys) + len(m.transaction.Data.Ownerships) + 2
		}
	}
}
