package domain

import "errors"

var (
	ErrEmptyOrderUID    = errors.New("empty order_uid")
	ErrEmptyTrackNumber = errors.New("empty track_number")
	ErrEmptyTransaction = errors.New("empty payment.transaction")
	ErrEmptyItems       = errors.New("empty items")
)