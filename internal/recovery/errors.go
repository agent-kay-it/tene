package recovery

import "errors"

var (
	ErrInvalidMnemonic = errors.New("recovery: invalid mnemonic phrase")
	ErrRecoveryFailed  = errors.New("recovery: master key recovery failed")
)
