package navigator

type PreflightCheck int

const (
	CheckWorkingDir PreflightCheck = iota
	CheckContainerEngine
	CheckPlaybook
	CheckNavigatorResolve
	CheckNavigatorBinary
)

type SetupStep int

const (
	SetupDir SetupStep = iota
	SetupPlaybook
	SetupInventories
	SetupExtraVars
	SetupPrivateKeys
	SetupKnownHosts
	SetupSettings
)

type runError struct {
	Message string
	Err     error
}

func (e runError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}

	return e.Message
}

func (e runError) Unwrap() error {
	return e.Err
}

type PreflightError struct {
	runError

	Check PreflightCheck
}

func newPreflightError(check PreflightCheck, message string, err error) *PreflightError {
	return &PreflightError{runError: runError{Message: message, Err: err}, Check: check}
}

type SetupError struct {
	runError

	Step SetupStep
}

func newSetupError(step SetupStep, message string, err error) *SetupError {
	return &SetupError{runError: runError{Message: message, Err: err}, Step: step}
}

type QueryError struct {
	runError

	Name string
}

func newQueryError(name string, message string, err error) *QueryError {
	return &QueryError{runError: runError{Message: message, Err: err}, Name: name}
}
