package connection

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/esxi-manager/esxi-manager/internal/config"
)

// Manager handles SSH connections with pooling and reconnection logic
type Manager struct {
	host        *config.ESXiHost
	client      *ssh.Client
	mu          sync.RWMutex
	maxAttempts int
	retryDelay  time.Duration
}

// NewManager creates a new connection manager for an ESXi host
func NewManager(host *config.ESXiHost) (*Manager, error) {
	if err := host.Validate(); err != nil {
		return nil, fmt.Errorf("invalid host configuration: %w", err)
	}

	return &Manager{
		host:        host,
		maxAttempts: 3,
		retryDelay:  2 * time.Second,
	}, nil
}

// Connect establishes an SSH connection to the ESXi host with retry logic
func (m *Manager) Connect() error {
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
func (m *Manager) GetClient() (*ssh.Client, error) {
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
	defer m.mu.RLock()
	return m.client, nil
}

// Close closes the SSH connection
func (m *Manager) Close() error {
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
func (m *Manager) dial() (*ssh.Client, error) {
	// Build authentication methods in order of preference
	authMethods := []ssh.AuthMethod{
		// Try password authentication first
		ssh.Password(m.host.Password),
		// Also try keyboard-interactive as fallback (many systems prefer this)
		ssh.KeyboardInteractive(m.keyboardInteractiveChallenge),
	}

	config := &ssh.ClientConfig{
		User:            m.host.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // ESXi hosts often have self-signed certs
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", m.host.URI, m.host.Port)
	return ssh.Dial("tcp", addr, config)
}

// keyboardInteractiveChallenge handles keyboard-interactive authentication
// This is needed for systems like ESXi that require interactive auth
func (m *Manager) keyboardInteractiveChallenge(user, instruction string, questions []string, echos []bool) ([]string, error) {
	// For keyboard-interactive, we respond to all prompts with the password
	// This handles scenarios where the server asks for password via interactive challenge
	answers := make([]string, len(questions))
	for i := range answers {
		answers[i] = m.host.Password
	}
	return answers, nil
}

// testConnection checks if the current connection is still alive
func (m *Manager) testConnection() error {
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
func (m *Manager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.client != nil && m.testConnection() == nil
}
