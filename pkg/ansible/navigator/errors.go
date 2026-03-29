package navigator

type PreflightCheckID int

const (
	CheckWorkingDir PreflightCheckID = iota
	CheckContainerEngine
	CheckPlaybook
	CheckNavigatorResolve
	CheckNavigatorBinary
)

type SetupComponentID int

const (
	SetupDir SetupComponentID = iota
	SetupPlaybook
	SetupInventories
	SetupExtraVars
	SetupPrivateKeys
	SetupKnownHosts
)

type PreflightError struct {
	Check   PreflightCheckID
	Message string
	Err     error
}

func (e *PreflightError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}

	return e.Message
}

func (e *PreflightError) Unwrap() error {
	return e.Err
}

type SetupError struct {
	Component SetupComponentID
	Message   string
	Err       error
}

func (e *SetupError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}

	return e.Message
}

func (e *SetupError) Unwrap() error {
	return e.Err
}
