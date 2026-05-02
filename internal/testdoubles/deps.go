package testdoubles

// ServerDeps aggregates all outbound-port fakes for unit tests.
// Add one field per outbound port as the inbound/outbound port split is formalised.
type ServerDeps struct{}

// NewServerDeps returns a ServerDeps with all fakes initialised to safe zero-value defaults.
func NewServerDeps() *ServerDeps {
	return &ServerDeps{}
}
