package navigator

type Mode int

const (
	ModeHost Mode = iota
	ModeEE
)

func (m Mode) String() string {
	switch m {
	case ModeHost:
		return "host"
	case ModeEE:
		return "ee"
	}

	return "unknown"
}

func (m Mode) UsesEE() bool {
	return m == ModeEE
}

func (c *RunConfig) mode() Mode {
	if c.Settings.EEEnabled {
		return ModeEE
	}

	return ModeHost
}
