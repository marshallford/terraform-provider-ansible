package ansible

import "slices"

type Status string

const (
	StatusUnknown    Status = "unknown"
	StatusSuccessful Status = "successful"
	StatusFailed     Status = "failed"
	StatusTimeout    Status = "timeout"
	StatusCanceled   Status = "canceled"
)

func validStatuses() []Status {
	return []Status{
		StatusSuccessful,
		StatusFailed,
		StatusTimeout,
		StatusCanceled,
	}
}

func ParseStatus(status string) Status {
	if slices.Contains(validStatuses(), Status(status)) {
		return Status(status)
	}

	return StatusUnknown
}
