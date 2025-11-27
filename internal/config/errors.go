package config

import "errors"

var (
	ErrMissingURI      = errors.New("esxi host URI is required")
	ErrMissingUsername = errors.New("esxi host username is required")
	ErrMissingPassword = errors.New("esxi host password is required")
)
