package fake

// LogCall represents a single log invocation for assertions.
type LogCall struct {
	Message string
	Args    []any
}

// FakeLogger is a test double that records all log calls for verification.
type MockLogger struct {
	InfoCalls  []LogCall
	ErrorCalls []LogCall
	WarnCalls  []LogCall
	DebugCalls []LogCall
}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (m *MockLogger) Info(msg string, args ...any) {
	m.InfoCalls = append(m.InfoCalls, LogCall{Message: msg, Args: args})
}

func (m *MockLogger) Error(msg string, args ...any) {
	m.ErrorCalls = append(m.ErrorCalls, LogCall{Message: msg, Args: args})
}

func (m *MockLogger) Warn(msg string, args ...any) {
	m.WarnCalls = append(m.WarnCalls, LogCall{Message: msg, Args: args})
}

func (m *MockLogger) Debug(msg string, args ...any) {
	m.DebugCalls = append(m.DebugCalls, LogCall{Message: msg, Args: args})
}

// Reset clears all recorded calls (useful between test cases).
func (m *MockLogger) Reset() {
	m.InfoCalls = nil
	m.ErrorCalls = nil
	m.WarnCalls = nil
	m.DebugCalls = nil
}
