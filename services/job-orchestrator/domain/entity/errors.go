package entity

import "errors"

var (
	ErrInvalidJobName    = errors.New("job name cannot be empty")
	ErrInvalidJarURL     = errors.New("jar URL cannot be empty")
	ErrInvalidEntryClass = errors.New("entry class cannot be empty")
	ErrJobNotFound       = errors.New("job not found")
	ErrJobAlreadyExists  = errors.New("job with this name already exists")
)

