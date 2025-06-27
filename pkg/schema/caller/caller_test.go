package caller_test

import (
	"errors"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xphir/terraform-buildkite-plugin/pkg/schema/caller"
)

func TestCaller_CallPath(t *testing.T) {
	tests := []struct {
		name          string
		caller        caller.Caller
		expectErr     bool
		expectErrText string
		expectPath    string
	}{
		{
			name: "Success with MockCaller",
			caller: &caller.MockCaller{
				CallPathResult: "./cmd",
				CallPathErr:    nil,
			},
			expectErr:  false,
			expectPath: "./cmd",
		},
		{
			name: "FindCallerError with MockCaller",
			caller: &caller.MockCaller{
				CallPathResult: "",
				CallPathErr:    errors.New("failed to find entrypoint caller"),
			},
			expectErr:     true,
			expectErrText: "failed to find entrypoint caller",
		},
		{
			name: "WorkingDirError with real implementation",
			caller: caller.New(
				caller.WithFindCallerFn(func(matcher func(frame runtime.Frame) bool) (runtime.Frame, error) {
					callStack := []runtime.Frame{
						{File: "/path/to/project/pkg/schema/caller_test.go"},
						{File: "/path/to/project/cmd/main.go"},
					}
					for _, frame := range callStack {
						if matcher(frame) {
							return frame, nil
						}
					}
					return runtime.Frame{}, errors.New("no matching frame found")
				}),
				caller.WithWorkingDirFn(func() (string, error) {
					return "", errors.New("failed to get working directory")
				}),
				caller.WithMatcherFn(func(f runtime.Frame) bool {
					return f.File == "/path/to/project/cmd/main.go"
				}),
			),
			expectErr:     true,
			expectErrText: "failed to get working directory",
		},
		{
			name: "RelativePathError with real implementation",
			caller: caller.New(
				caller.WithFindCallerFn(func(_ func(frame runtime.Frame) bool) (runtime.Frame, error) {
					return runtime.Frame{File: "nowhere.go"}, nil
				}),
				caller.WithWorkingDirFn(func() (string, error) {
					return "/path/to/project", nil
				}),
				caller.WithMatcherFn(func(_ runtime.Frame) bool { return true }),
			),
			expectErr:     true,
			expectErrText: "failed to make path relative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			relPath, err := tt.caller.CallPath()
			if tt.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectErrText)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectPath, relPath)
			}
		})
	}
}

func TestCaller_DefaultConstructor(t *testing.T) {
	c := caller.New()
	require.NotNil(t, c)
}
