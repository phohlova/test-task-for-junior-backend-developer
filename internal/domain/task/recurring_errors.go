package task

import "errors"

var (
	ErrEmptyType               = errors.New("recurrence type cannot be empty")
	ErrZeroStartDate           = errors.New("start date cannot be zero")
	ErrEndDateBeforeStart      = errors.New("end date must be after or equal to start date")
	ErrInvalidTimezone         = errors.New("invalid timezone")
	ErrInvalidInterval         = errors.New("daily interval must be >= 1")
	ErrInvalidDayOfMonth       = errors.New("day of month must be between 1 and 31")
	ErrInvalidOverflowStrategy = errors.New("invalid overflow strategy")
	ErrEmptyDates              = errors.New("specific dates list cannot be empty")
	ErrDuplicateDate           = errors.New("specific dates list contains duplicates")
	ErrZeroDateInList          = errors.New("specific dates list contains zero time")
	ErrInvalidParityMode       = errors.New("parity mode must be 'even' or 'odd'")
	ErrUnknownType             = errors.New("unknown recurrence type")
)