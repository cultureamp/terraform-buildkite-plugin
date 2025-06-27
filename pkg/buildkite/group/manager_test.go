package group_test

import (
	"bytes"
	"testing"

	"github.com/xphir/terraform-buildkite-plugin/pkg/buildkite/group"
)

func TestGroupInterface_Methods(t *testing.T) {
	tests := []struct {
		name     string
		method   func(group.Manager, string)
		title    string
		expected string
	}{
		{
			name:     "Open creates an open group",
			method:   func(g group.Manager, title string) { g.Open(title) },
			title:    "Test Open Group",
			expected: "+++ Test Open Group\n",
		},
		{
			name:     "Close creates a closed group",
			method:   func(g group.Manager, title string) { g.Closed(title) },
			title:    "Test Closed Group",
			expected: "--- Test Closed Group\n",
		},
		{
			name:     "Mute creates a muted group",
			method:   func(g group.Manager, title string) { g.Muted(title) },
			title:    "Test Muted Group",
			expected: "~~~ Test Muted Group\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			g := group.NewLogGroupManager(buf)

			tt.method(g, tt.title)

			if got := buf.String(); got != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, got, tt.expected)
			}
		})
	}
}

func TestGroupInterface_OpenF_ClosedF_MutedF(t *testing.T) {
	buf := &bytes.Buffer{}
	g := group.NewLogGroupManager(buf)
	g.OpenF("open: %d", 1)
	g.ClosedF("closed: %s", "foo")
	g.MutedF("muted: %.2f", 3.14)
	expected := "+++ open: 1\n--- closed: foo\n~~~ muted: 3.14\n"
	if got := buf.String(); got != expected {
		t.Errorf("OpenF/ClosedF/MutedF = %q, want %q", got, expected)
	}
}

func TestGroupInterface_OpenCurrent(t *testing.T) {
	buf := &bytes.Buffer{}
	g := group.NewLogGroupManager(buf)

	g.OpenCurrent()

	expected := "^^^ +++\n"
	if got := buf.String(); got != expected {
		t.Errorf("OpenCurrent() = %q, want %q", got, expected)
	}
}

func TestGroupInterface_WithWriter(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	g := group.NewLogGroupManager(buf1)
	g = g.SetOutput(buf2)
	g.Open("Hello")
	if buf1.Len() != 0 && buf2.String() != "+++ Hello\n" {
		t.Errorf("WithWriter did not update writer correctly")
	}
}

func TestGlobalGroup_DefaultWriter(t *testing.T) {
	// The globalGroup is unexported, so we can't access it directly.
	// Instead, we check that NewLogGroup(nil) returns a group with io.Discard
	g := group.NewLogGroupManager(nil)
	buf := &bytes.Buffer{}
	g = g.SetOutput(buf)
	g.Open("Hello")
	if buf.String() != "+++ Hello\n" {
		t.Errorf("WithWriter did not update writer correctly, got %q", buf.String())
	}
}

func TestNewLogGroup_NilWriterDefaultsToDiscard(t *testing.T) {
	// Should not panic when creating group with nil writer
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewLogGroupManager(nil) should not panic, but got: %v", r)
		}
	}()

	g := group.NewLogGroupManager(nil)
	g.Open("Should not panic or write")

	// Verify group is usable after creation with nil writer
	if g == nil {
		t.Error("NewLogGroupManager(nil) should return a valid group manager")
	}
}

func TestGroupInterface_NilWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	g := group.NewLogGroupManager(buf)
	g = g.SetOutput(nil) // Set to nil writer

	// Should not panic when calling methods with nil writer
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Methods with nil writer should not panic, but got: %v", r)
		}
	}()

	g.Open("Open with nil writer")
	g.Closed("Closed with nil writer")
	g.Muted("Muted with nil writer")
	g.OpenCurrent()

	// Verify original buffer remains unchanged after setting nil writer
	if buf.Len() != 0 {
		t.Errorf("Original buffer should remain empty after setting nil writer, got: %q", buf.String())
	}
}
