package config

// ESXiHost represents the connection parameters for an ESXi host
type ESXiHost struct {
	URI      string
	Username string
	Password string
	Port     int
}

// Validate checks if the ESXi host configuration is valid
func (e *ESXiHost) Validate() error {
	if e.URI == "" {
		return ErrMissingURI
	}
	if e.Username == "" {
		return ErrMissingUsername
	}
	if e.Password == "" {
		return ErrMissingPassword
	}
	if e.Port == 0 {
		e.Port = 22 // Default SSH port
	}
	return nil
}
