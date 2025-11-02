package usecase

import "errors"

var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrInternal            = errors.New("internal error")
	ErrForbidden           = errors.New("action forbidden")
	ErrFailedToSaveBanner  = errors.New("failed to save banner")
	ErrFailedToCreateSpool = errors.New("failed to create spool")
	ErrFailedToUpdateSpool = errors.New("failed to update spool")
	ErrFailedToInvite      = errors.New("failed to invite member to spool")
	ErrFailedToGetMembers  = errors.New("failed to get spool members")
	ErrFailedToGetSpool    = errors.New("failed to get spool info")
)
