package utils

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/esxi-manager/esxi-manager/internal/esxi/common"
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
)

var (
	ErrNotConnected = errors.New("not connected to ESXi host")
	ErrDialFailed   = errors.New("failed to establish SSH connection")
)

// ValidateConnectionParams validates that all required connection parameters are set
// This should be called before attempting to create an SSHManager
func ValidateConnectionParams(params *config.Params) error {
	if params.ESXiHostURI == "" {
		return common.NewValidationError("esxi-host-uri", "is required")
	}
	if params.ESXiHostUsername == "" {
		return common.NewValidationError("esxi-host-username", "is required")
	}
	if params.ESXiHostPassword == "" {
		return common.NewValidationError("esxi-host-password", "is required")
	}
	if params.ESXiHostPort <= 0 {
		return common.NewValidationErrorWithValue("esxi-host-port", "must be greater than 0", fmt.Sprintf("%d", params.ESXiHostPort))
	}
	if params.ESXiHostPort > 65535 {
		return common.NewValidationErrorWithValue("esxi-host-port", "must be less than or equal to 65535", fmt.Sprintf("%d", params.ESXiHostPort))
	}
	return nil
}

// SSHManager handles SSH connections with pooling and reconnection logic
type SSHManager struct {
	params      *config.Params
	client      *ssh.Client
	mu          sync.RWMutex
	maxAttempts int
	retryDelay  time.Duration
}

// NewSSHManager creates a new SSH connection manager for an ESXi host
// Validates connection parameters and returns any validation errors
func NewSSHManager(params *config.Params) (*SSHManager, error) {
	if err := ValidateConnectionParams(params); err != nil {
		return nil, err
	}
	return &SSHManager{
		params:      params,
		maxAttempts: 3,
		retryDelay:  2 * time.Second,
	}, nil
}

// Connect establishes an SSH connection to the ESXi host with retry logic
func (m *SSHManager) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If already connected, return
	if m.client != nil {
		if err := m.testConnection(); err == nil {
			return nil
		}
		// Connection is stale, close it
		m.client.Close()
		m.client = nil
	}

	// Try to connect with retry logic
	var lastErr error
	for attempt := 1; attempt <= m.maxAttempts; attempt++ {
		client, err := m.dial()
		if err == nil {
			m.client = client
			return nil
		}

		lastErr = err
		if attempt < m.maxAttempts {
			time.Sleep(m.retryDelay)
		}
	}

	return fmt.Errorf("failed to connect after %d attempts: %w", m.maxAttempts, lastErr)
}

// GetClient returns the active SSH client, reconnecting if necessary
func (m *SSHManager) GetClient() (*ssh.Client, error) {
	m.mu.RLock()
	if m.client != nil {
		if err := m.testConnection(); err == nil {
			defer m.mu.RUnlock()
			return m.client, nil
		}
	}
	m.mu.RUnlock()

	// Connection is stale or doesn't exist, reconnect
	if err := m.Connect(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.client, nil
}

// Close closes the SSH connection
func (m *SSHManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client != nil {
		err := m.client.Close()
		m.client = nil
		return err
	}

	return nil
}

// dial creates a new SSH connection to the ESXi host
func (m *SSHManager) dial() (*ssh.Client, error) {
	// Build authentication methods in order of preference
	authMethods := []ssh.AuthMethod{
		// Try password authentication first
		ssh.Password(m.params.ESXiHostPassword),
		// Also try keyboard-interactive as fallback (many systems prefer this)
		ssh.KeyboardInteractive(m.keyboardInteractiveChallenge),
	}

	config := &ssh.ClientConfig{
		User:            m.params.ESXiHostUsername,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // ESXi hosts often have self-signed certs
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", m.params.ESXiHostURI, m.params.ESXiHostPort)
	return ssh.Dial("tcp", addr, config)
}

// keyboardInteractiveChallenge handles keyboard-interactive authentication
// This is needed for systems like ESXi that require interactive auth
func (m *SSHManager) keyboardInteractiveChallenge(user, instruction string, questions []string, echos []bool) ([]string, error) {
	// For keyboard-interactive, we respond to all prompts with the password
	// This handles scenarios where the server asks for password via interactive challenge
	answers := make([]string, len(questions))
	for i := range answers {
		answers[i] = m.params.ESXiHostPassword
	}
	return answers, nil
}

// testConnection checks if the current connection is still alive
func (m *SSHManager) testConnection() error {
	if m.client == nil {
		return ErrNotConnected
	}

	// Send a simple command to test connectivity
	session, err := m.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Simple echo command to verify connection
	if err := session.Run("echo 'test'"); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	return nil
}

// IsConnected returns true if there is an active connection
func (m *SSHManager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.client != nil && m.testConnection() == nil
}

// RunCommand executes a command and returns its output
func (m *SSHManager) RunCommand(cmd string) (string, error) {
	if err := m.Connect(); err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}

	client, err := m.GetClient()
	if err != nil {
		return "", fmt.Errorf("failed to get client: %w", err)
	}

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.Output(cmd)
	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}
