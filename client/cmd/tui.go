package main

import (
	"context"
	"fmt"
	"strings"

	"data-vault/client/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// sessionState represents the current state of the TUI application
type sessionState int

const (
	mainMenuView sessionState = iota
	loginView
	registerView
	dataMenuView
	postDataView
	getDataView
	deleteDataView
	pingView
)

// model represents the complete TUI application state
type model struct {
	state      sessionState
	choices    []string
	cursor     int
	selected   map[int]struct{}
	username   string
	password   string
	data       string
	dataType   string
	dataID     string
	jwtToken   string
	message    string
	inputMode  bool
	inputField string
	userData   []models.Data
	err        error
}

// initialModel creates and returns the initial TUI model
func initialModel() model {
	return model{
		state:     mainMenuView,
		choices:   []string{"Login", "Register", "Ping Server", "Quit"},
		selected:  make(map[int]struct{}),
		inputMode: false,
	}
}

// Init initializes the TUI application
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles all state updates for the TUI
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case mainMenuView:
			return m.updateMainMenu(msg)
		case loginView:
			return m.updateLogin(msg)
		case registerView:
			return m.updateRegister(msg)
		case dataMenuView:
			return m.updateDataMenu(msg)
		case postDataView:
			return m.updatePostData(msg)
		case getDataView:
			return m.updateGetData(msg)
		case deleteDataView:
			return m.updateDeleteData(msg)
		case pingView:
			return m.updatePing(msg)
		}
	case loginMsg:
		if msg.success {
			m.jwtToken = msg.token
			m.message = "Login successful! JWT token received."
			m.state = dataMenuView
			m.cursor = 0
		} else {
			m.message = fmt.Sprintf("Login failed: %v", msg.err)
		}
		m.resetInput()
	case registerMsg:
		if msg.success {
			m.jwtToken = msg.token
			m.message = "Registration successful! JWT token received."
			m.state = dataMenuView
			m.cursor = 0
		} else {
			m.message = fmt.Sprintf("Registration failed: %v", msg.err)
		}
		m.resetInput()
	case postDataMsg:
		if msg.success {
			m.message = "Data posted successfully!"
		} else {
			m.message = fmt.Sprintf("Failed to post data: %v", msg.err)
		}
		m.state = dataMenuView
		m.cursor = 0
		m.resetInput()
	case getDataMsg:
		if msg.err != nil {
			m.message = fmt.Sprintf("Failed to get data: %v", msg.err)
		} else {
			m.userData = msg.data
			if len(msg.data) == 0 {
				m.message = "No data found."
			} else {
				m.message = fmt.Sprintf("Retrieved %d data items.", len(msg.data))
			}
		}
	case deleteDataMsg:
		if msg.success {
			m.message = "Data deleted successfully!"
		} else {
			m.message = fmt.Sprintf("Failed to delete data: %v", msg.err)
		}
		m.state = dataMenuView
		m.cursor = 0
		m.resetInput()
	case pingMsg:
		if msg.success {
			m.message = "✓ Server is reachable!"
		} else {
			m.message = "✗ Server is not reachable!"
		}
	}
	return m, nil
}

// updateMainMenu handles main menu navigation
func (m model) updateMainMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.choices)-1 {
			m.cursor++
		}
	case "enter", " ":
		switch m.cursor {
		case 0:
			m.state = loginView
			m.inputMode = true
			m.inputField = "username"
			m.message = ""
		case 1:
			m.state = registerView
			m.inputMode = true
			m.inputField = "username"
			m.message = ""
		case 2:
			m.state = pingView
			return m, m.pingServerCmd()
		case 3:
			return m, tea.Quit
		}
	}
	return m, nil
}

// updateLogin handles login form input
func (m model) updateLogin(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.state = mainMenuView
		m.cursor = 0
		m.resetInput()
	case "enter":
		if m.inputField == "username" && m.username != "" {
			m.inputField = "password"
		} else if m.inputField == "password" && m.password != "" {
			return m, m.loginCmd()
		}
	case "backspace":
		if m.inputField == "username" && len(m.username) > 0 {
			m.username = m.username[:len(m.username)-1]
		} else if m.inputField == "password" && len(m.password) > 0 {
			m.password = m.password[:len(m.password)-1]
		}
	default:
		if len(msg.String()) == 1 {
			if m.inputField == "username" {
				m.username += msg.String()
			} else if m.inputField == "password" {
				m.password += msg.String()
			}
		}
	}
	return m, nil
}

// updateRegister handles registration form input
func (m model) updateRegister(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.state = mainMenuView
		m.cursor = 0
		m.resetInput()
	case "enter":
		if m.inputField == "username" && m.username != "" {
			m.inputField = "password"
		} else if m.inputField == "password" && m.password != "" {
			return m, m.registerCmd()
		}
	case "backspace":
		if m.inputField == "username" && len(m.username) > 0 {
			m.username = m.username[:len(m.username)-1]
		} else if m.inputField == "password" && len(m.password) > 0 {
			m.password = m.password[:len(m.password)-1]
		}
	default:
		if len(msg.String()) == 1 {
			if m.inputField == "username" {
				m.username += msg.String()
			} else if m.inputField == "password" {
				m.password += msg.String()
			}
		}
	}
	return m, nil
}

// updateDataMenu handles data operations menu navigation
func (m model) updateDataMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	dataChoices := []string{"Post Data", "Get Data", "Delete Data", "Back to Main Menu"}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.state = mainMenuView
		m.cursor = 0
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(dataChoices)-1 {
			m.cursor++
		}
	case "enter", " ":
		switch m.cursor {
		case 0:
			m.state = postDataView
			m.inputMode = true
			m.inputField = "data"
			m.message = ""
		case 1:
			m.state = getDataView
			return m, m.getDataCmd()
		case 2:
			m.state = deleteDataView
			m.inputMode = true
			m.inputField = "dataID"
			m.message = ""
		case 3:
			m.state = mainMenuView
			m.cursor = 0
		}
	}
	return m, nil
}

// updatePostData handles post data form input
func (m model) updatePostData(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.state = dataMenuView
		m.cursor = 0
		m.resetInput()
	case "enter":
		if m.data != "" {
			return m, m.postDataCmd()
		}
	case "backspace":
		if len(m.data) > 0 {
			m.data = m.data[:len(m.data)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.data += msg.String()
		}
	}
	return m, nil
}

// updateGetData handles get data view navigation
func (m model) updateGetData(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "enter":
		m.state = dataMenuView
		m.cursor = 0
	}
	return m, nil
}

// updateDeleteData handles delete data form input
func (m model) updateDeleteData(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.state = dataMenuView
		m.cursor = 0
		m.resetInput()
	case "enter":
		if m.dataID != "" {
			return m, m.deleteDataCmd()
		}
	case "backspace":
		if len(m.dataID) > 0 {
			m.dataID = m.dataID[:len(m.dataID)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.dataID += msg.String()
		}
	}
	return m, nil
}

// updatePing handles ping view navigation
func (m model) updatePing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "enter":
		m.state = mainMenuView
		m.cursor = 0
	}
	return m, nil
}

// resetInput clears all input fields and resets input state
func (m model) resetInput() {
	m.username = ""
	m.password = ""
	m.data = ""
	m.dataType = ""
	m.dataID = ""
	m.inputMode = false
	m.inputField = ""
	m.message = ""
}

// loginMsg represents the result of a login operation
type loginMsg struct {
	success bool
	token   string
	err     error
}

// registerMsg represents the result of a registration operation
type registerMsg struct {
	success bool
	token   string
	err     error
}

// postDataMsg represents the result of a post data operation
type postDataMsg struct {
	success bool
	err     error
}

// getDataMsg represents the result of a get data operation
type getDataMsg struct {
	data []models.Data
	err  error
}

// deleteDataMsg represents the result of a delete data operation
type deleteDataMsg struct {
	success bool
	err     error
}

// pingMsg represents the result of a server ping operation
type pingMsg struct {
	success bool
	err     error
}

// loginCmd creates a command to perform user login
func (m model) loginCmd() tea.Cmd {
	return func() tea.Msg {
		service, err := initService()
		if err != nil {
			return loginMsg{success: false, err: err}
		}

		user := models.User{Login: m.username, Password: m.password}
		jwt, err := service.Login(context.Background(), user)
		if err != nil {
			return loginMsg{success: false, err: err}
		}

		return loginMsg{success: true, token: jwt}
	}
}

// registerCmd creates a command to register a new user
func (m model) registerCmd() tea.Cmd {
	return func() tea.Msg {
		service, err := initService()
		if err != nil {
			return registerMsg{success: false, err: err}
		}

		user := models.User{Login: m.username, Password: m.password}
		jwt, err := service.Register(context.Background(), user)
		if err != nil {
			return registerMsg{success: false, err: err}
		}

		return registerMsg{success: true, token: jwt}
	}
}

// postDataCmd creates a command to post data to the server
func (m model) postDataCmd() tea.Cmd {
	return func() tea.Msg {
		service, err := initService()
		if err != nil {
			return postDataMsg{success: false, err: err}
		}

		dataType := m.dataType
		if dataType == "" {
			dataType = "text"
		}

		err = service.PostData(context.Background(), m.jwtToken, dataType, []byte(m.data))
		if err != nil {
			return postDataMsg{success: false, err: err}
		}

		return postDataMsg{success: true}
	}
}

// getDataCmd creates a command to retrieve data from the server
func (m model) getDataCmd() tea.Cmd {
	return func() tea.Msg {
		service, err := initService()
		if err != nil {
			return getDataMsg{err: err}
		}

		data, err := service.GetData(context.Background(), m.jwtToken)
		if err != nil {
			return getDataMsg{err: err}
		}

		return getDataMsg{data: data}
	}
}

// deleteDataCmd creates a command to delete data from the server
func (m model) deleteDataCmd() tea.Cmd {
	return func() tea.Msg {
		service, err := initService()
		if err != nil {
			return deleteDataMsg{success: false, err: err}
		}

		err = service.DeleteData(context.Background(), m.jwtToken, m.dataID)
		if err != nil {
			return deleteDataMsg{success: false, err: err}
		}

		return deleteDataMsg{success: true}
	}
}

// pingServerCmd creates a command to ping the server
func (m model) pingServerCmd() tea.Cmd {
	return func() tea.Msg {
		service, err := initService()
		if err != nil {
			return pingMsg{success: false, err: err}
		}

		success := service.PingServer(context.Background())
		return pingMsg{success: success}
	}
}

// TUI styling variables
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F56")).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)
)

// View renders the current state of the TUI
func (m model) View() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("Data Vault Client"))
	s.WriteString("\n\n")

	// Show JWT status if logged in
	if m.jwtToken != "" {
		s.WriteString(messageStyle.Render("✓ Authenticated"))
		s.WriteString("\n\n")
	}

	switch m.state {
	case mainMenuView:
		s.WriteString("Choose an option:\n\n")
		for i, choice := range m.choices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
				choice = selectedStyle.Render(choice)
			}
			s.WriteString(fmt.Sprintf("%s %s\n", cursor, choice))
		}

	case loginView:
		s.WriteString("Login\n\n")
		s.WriteString(fmt.Sprintf("Username: %s\n", inputStyle.Render(m.username)))
		if m.inputField == "username" {
			s.WriteString("█")
		}
		s.WriteString("\n")

		passwordDisplay := strings.Repeat("*", len(m.password))
		s.WriteString(fmt.Sprintf("Password: %s\n", inputStyle.Render(passwordDisplay)))
		if m.inputField == "password" {
			s.WriteString("█")
		}
		s.WriteString("\n\nPress Enter to continue, Esc to go back")

	case registerView:
		s.WriteString("Register\n\n")
		s.WriteString(fmt.Sprintf("Username: %s\n", inputStyle.Render(m.username)))
		if m.inputField == "username" {
			s.WriteString("█")
		}
		s.WriteString("\n")

		passwordDisplay := strings.Repeat("*", len(m.password))
		s.WriteString(fmt.Sprintf("Password: %s\n", inputStyle.Render(passwordDisplay)))
		if m.inputField == "password" {
			s.WriteString("█")
		}
		s.WriteString("\n\nPress Enter to continue, Esc to go back")

	case dataMenuView:
		if m.jwtToken == "" {
			s.WriteString(errorStyle.Render("Please login first!"))
			s.WriteString("\n\nPress Esc to go back to main menu")
		} else {
			s.WriteString("Data Operations:\n\n")
			dataChoices := []string{"Post Data", "Get Data", "Delete Data", "Back to Main Menu"}
			for i, choice := range dataChoices {
				cursor := " "
				if m.cursor == i {
					cursor = ">"
					choice = selectedStyle.Render(choice)
				}
				s.WriteString(fmt.Sprintf("%s %s\n", cursor, choice))
			}
		}

	case postDataView:
		s.WriteString("Post Data\n\n")
		s.WriteString(fmt.Sprintf("Data: %s\n", inputStyle.Render(m.data)))
		s.WriteString("█")
		s.WriteString("\n\nPress Enter to submit, Esc to go back")

	case getDataView:
		s.WriteString("Your Data:\n\n")
		if len(m.userData) == 0 {
			s.WriteString("No data found.")
		} else {
			for i, item := range m.userData {
				s.WriteString(fmt.Sprintf("%d. ID: %s\n", i+1, item.ID))
				s.WriteString(fmt.Sprintf("   Type: %s\n", item.Type))
				s.WriteString(fmt.Sprintf("   Data: %s\n", string(item.Data)))
				s.WriteString(fmt.Sprintf("   Uploaded: %s\n\n", item.UploadedAt))
			}
		}
		s.WriteString("\nPress Enter or Esc to go back")

	case deleteDataView:
		s.WriteString("Delete Data\n\n")
		s.WriteString(fmt.Sprintf("Data ID: %s\n", inputStyle.Render(m.dataID)))
		s.WriteString("█")
		s.WriteString("\n\nPress Enter to delete, Esc to go back")

	case pingView:
		s.WriteString("Server Status\n\n")
		s.WriteString("Checking server connectivity...")
		s.WriteString("\n\nPress Enter or Esc to go back")
	}

	// Show message if any
	if m.message != "" {
		s.WriteString("\n\n")
		if strings.Contains(m.message, "failed") || strings.Contains(m.message, "Error") {
			s.WriteString(errorStyle.Render(m.message))
		} else {
			s.WriteString(messageStyle.Render(m.message))
		}
	}

	s.WriteString("\n\n")
	s.WriteString("Press q or ctrl+c to quit")

	return s.String()
}
