package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintJSONAddsNewline(t *testing.T) {
	var buf bytes.Buffer
	if err := PrintJSON(&buf, map[string]string{"hello": "world"}); err != nil {
		t.Fatalf("PrintJSON() error = %v", err)
	}
	if !strings.Contains(buf.String(), "\"hello\": \"world\"") {
		t.Fatalf("json output = %q, want hello field", buf.String())
	}
	if !strings.HasSuffix(buf.String(), "\n") {
		t.Fatalf("json output = %q, want trailing newline", buf.String())
	}
}
