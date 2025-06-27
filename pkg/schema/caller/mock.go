package caller

// MockCaller is a test double for Caller.
type MockCaller struct {
	CallPathFunc   func() (string, error)
	CallPathResult string
	CallPathErr    error
}

func (m *MockCaller) CallPath() (string, error) {
	if m.CallPathFunc != nil {
		return m.CallPathFunc()
	}
	return m.CallPathResult, m.CallPathErr
}
