package provider

import "errors"

var (
	ErrNotSupported  = errors.New("provider: not supported")
	ErrNotFound      = errors.New("provider: not found")
	ErrUnauthorized  = errors.New("provider: unauthorized")
	ErrOffline       = errors.New("provider: offline")
	ErrRateLimited   = errors.New("provider: rate limited")
	ErrTemporary     = errors.New("provider: temporary failure")
	ErrInvalidConfig = errors.New("provider: invalid config")
)

func IsNotSupported(err error) bool { return errors.Is(err, ErrNotSupported) }
func IsNotFound(err error) bool     { return errors.Is(err, ErrNotFound) }
func IsUnauthorized(err error) bool { return errors.Is(err, ErrUnauthorized) }
func IsOffline(err error) bool      { return errors.Is(err, ErrOffline) }
func IsRateLimited(err error) bool  { return errors.Is(err, ErrRateLimited) }
func IsTemporary(err error) bool    { return errors.Is(err, ErrTemporary) }
