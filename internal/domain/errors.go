package domain

import "errors"

var (
	ErrGroupNotFound      = errors.New("group not found")
	ErrGroupAlreadyExists = errors.New("group already exists")
	ErrInvalidGroupID     = errors.New("invalid group id")
	ErrIPNotFound         = errors.New("ip not found")
	ErrIPAlreadyExists    = errors.New("ip already exists")
	ErrInvalidDateFormat  = errors.New("invalid date format")
)
