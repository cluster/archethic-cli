package tui

import (
	"fmt"
	"log"
	"os"

	"github.com/archethic-foundation/archethic-cli/tui/generateaddressui"
	"github.com/archethic-foundation/archethic-cli/tui/keychaincreatetransactionui"
	"github.com/archethic-foundation/archethic-cli/tui/keychainmanagementui"
	"github.com/archethic-foundation/archethic-cli/tui/mainui"
	tea "github.com/charmbracelet/bubbletea"
)

var p *tea.Program

type sessionState int

const (
	menuView sessionState = iota
	generateAddressView
	keychainManagementView
	keychainCreateTransactionView
	monthView
	loadingView
)

type MainModel struct {
	state                     sessionState
	main                      tea.Model
	generateAddress           tea.Model
	keychainManagement        tea.Model
	keychainCreateTransaction tea.Model
	ActiveMenuID              uint
	windowSize                tea.WindowSizeMsg
}

// StartTea the entry point for the UI. Initializes the model.
func StartTea(pvKeyBytes []byte) {
	if f, err := tea.LogToFile("debug.log", "help"); err != nil {
		fmt.Println("Couldn't open a file for logging:", err)
		os.Exit(1)
	} else {
		defer func() {
			err = f.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	m := New(pvKeyBytes)
	p = tea.NewProgram(m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// New initialize the main model for your program
func New(pvKeyBytes []byte) MainModel {
	return MainModel{
		state:                     sessionState(0),
		main:                      mainui.New(),
		generateAddress:           generateaddressui.New(),
		keychainManagement:        keychainmanagementui.New(pvKeyBytes),
		keychainCreateTransaction: keychaincreatetransactionui.New(pvKeyBytes),
	}
}

// Init run any intial IO on program start
func (m MainModel) Init() tea.Cmd {
	return nil
}

// Update handle IO and commands
func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		// Update the sub views when the size of the window is changed
		// It appears this is called during the init phase to get the inital
		// window size.
		m.generateAddress, _ = m.generateAddress.Update(msg)
		m.main, _ = m.main.Update(msg)
	case generateaddressui.BackMsg:
		m.state = menuView
	case keychainmanagementui.BackMsg:
		m.state = menuView
	case keychaincreatetransactionui.BackMsg:
		m.state = menuView
	case keychaincreatetransactionui.CreateTransactionMsg:
		m.state = keychainCreateTransactionView
	case mainui.SelectMsg:
		switch msg.ActiveMenu {
		case 1:
			m.state = generateAddressView
		case 2:
			m.state = keychainCreateTransactionView
		case 3:
			m.state = keychainManagementView
		}
	}

	switch m.state {
	case menuView:
		newProject, newCmd := m.main.Update(msg)
		newModel, ok := newProject.(mainui.Model)
		if !ok {
			panic("could not perform assertion on mainui model")
		}
		m.main = newModel
		cmd = newCmd
	case generateAddressView:
		newGenerateAddress, newCmd := m.generateAddress.Update(msg)
		newModel, ok := newGenerateAddress.(generateaddressui.Model)
		if !ok {
			panic("could not perform assertion on generateaddressui model")
		}
		m.generateAddress = newModel
		cmd = newCmd
	case keychainManagementView:
		newKeychainManagement, newCmd := m.keychainManagement.Update(msg)
		newModel, ok := newKeychainManagement.(keychainmanagementui.Model)
		if !ok {
			panic("could not perform assertion on keychainmanagement model")
		}
		if !newModel.IsInit {
			cmds = append(cmds, newModel.Init())
			newModel.IsInit = true
		}
		m.keychainManagement = newModel
		cmd = newCmd
	case keychainCreateTransactionView:
		newKeychainCreateTransaction, newCmd := m.keychainCreateTransaction.Update(msg)
		newModel, ok := newKeychainCreateTransaction.(keychaincreatetransactionui.Model)
		if !ok {
			panic("could not perform assertion on keychainmanagement model")
		}
		if !newModel.IsInit {
			cmds = append(cmds, newModel.Init())
			newModel.IsInit = true
		}
		m.keychainCreateTransaction = newModel
		cmd = newCmd
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// View return the text UI to be output to the terminal
func (m MainModel) View() string {
	switch m.state {
	case menuView:
		return m.main.View()
	case generateAddressView:
		return m.generateAddress.View()
	case keychainManagementView:
		return m.keychainManagement.View()
	case keychainCreateTransactionView:
		return m.keychainCreateTransaction.View()
	default:
		return m.main.View()
	}
}
