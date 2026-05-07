package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestSelectRegion_ChoosesTickTick(t *testing.T) {
	streams, stdout, _ := newTestStreams()
	streams.In = bytes.NewBufferString("1\n")

	region, err := SelectRegion(streams, "")
	if err != nil {
		t.Fatalf("SelectRegion() error = %v", err)
	}
	if region != "ticktick" {
		t.Fatalf("region = %q, want ticktick", region)
	}

	out := stdout.String()
	if !strings.Contains(out, "1. ticktick") {
		t.Fatalf("output missing ticktick option: %q", out)
	}
	if !strings.Contains(out, "Select:") {
		t.Fatalf("output missing prompt: %q", out)
	}
}

func TestSelectRegion_DefaultsToExistingRegion(t *testing.T) {
	streams, stdout, _ := newTestStreams()
	streams.In = bytes.NewBufferString("\n")

	region, err := SelectRegion(streams, "dida365")
	if err != nil {
		t.Fatalf("SelectRegion() error = %v", err)
	}
	if region != "dida365" {
		t.Fatalf("region = %q, want dida365", region)
	}

	out := stdout.String()
	if !strings.Contains(out, "Select [2]:") {
		t.Fatalf("output missing default prompt: %q", out)
	}
}

func TestSelectRegion_InvalidThenValid(t *testing.T) {
	streams, stdout, _ := newTestStreams()
	streams.In = bytes.NewBufferString("abc\n3\n2\n")

	region, err := SelectRegion(streams, "")
	if err != nil {
		t.Fatalf("SelectRegion() error = %v", err)
	}
	if region != "dida365" {
		t.Fatalf("region = %q, want dida365", region)
	}

	out := stdout.String()
	if strings.Count(out, "Invalid selection") != 2 {
		t.Fatalf("expected 2 invalid selections, got %d in output: %q", strings.Count(out, "Invalid selection"), out)
	}
}

func TestPrompt_WithDefault(t *testing.T) {
	streams, stdout, _ := newTestStreams()
	streams.In = bytes.NewBufferString("\n")

	result, err := Prompt(streams, "Enter name", "default")
	if err != nil {
		t.Fatalf("Prompt() error = %v", err)
	}
	if result != "default" {
		t.Fatalf("result = %q, want default", result)
	}

	out := stdout.String()
	if !strings.Contains(out, "Enter name [default]:") {
		t.Fatalf("output missing prompt with default: %q", out)
	}
}

func TestPrompt_OverridesDefault(t *testing.T) {
	streams, stdout, _ := newTestStreams()
	streams.In = bytes.NewBufferString("override\n")

	result, err := Prompt(streams, "Enter name", "default")
	if err != nil {
		t.Fatalf("Prompt() error = %v", err)
	}
	if result != "override" {
		t.Fatalf("result = %q, want override", result)
	}

	out := stdout.String()
	if !strings.Contains(out, "Enter name [default]:") {
		t.Fatalf("output missing prompt with default: %q", out)
	}
}

func TestPrompt_NoDefault(t *testing.T) {
	streams, stdout, _ := newTestStreams()
	streams.In = bytes.NewBufferString("value\n")

	result, err := Prompt(streams, "Enter name", "")
	if err != nil {
		t.Fatalf("Prompt() error = %v", err)
	}
	if result != "value" {
		t.Fatalf("result = %q, want value", result)
	}

	out := stdout.String()
	if !strings.Contains(out, "Enter name:") {
		t.Fatalf("output missing prompt without default: %q", out)
	}
}

func TestPromptSecret_KeepExisting(t *testing.T) {
	streams, stdout, _ := newTestStreams()
	streams.In = bytes.NewBufferString("\n")

	result, err := PromptSecret(streams, "API Key", true)
	if err != nil {
		t.Fatalf("PromptSecret() error = %v", err)
	}
	if result != "" {
		t.Fatalf("result = %q, want empty string", result)
	}

	out := stdout.String()
	if !strings.Contains(out, "API Key (press Enter to keep existing):") {
		t.Fatalf("output missing keep-existing prompt: %q", out)
	}
}

func TestPromptSecret_EnterNew(t *testing.T) {
	streams, stdout, _ := newTestStreams()
	streams.In = bytes.NewBufferString("new-secret\n")

	result, err := PromptSecret(streams, "API Key", true)
	if err != nil {
		t.Fatalf("PromptSecret() error = %v", err)
	}
	if result != "new-secret" {
		t.Fatalf("result = %q, want new-secret", result)
	}

	out := stdout.String()
	if !strings.Contains(out, "API Key (press Enter to keep existing):") {
		t.Fatalf("output missing keep-existing prompt: %q", out)
	}
}

func TestPromptSecret_NoExisting(t *testing.T) {
	streams, stdout, _ := newTestStreams()
	streams.In = bytes.NewBufferString("secret\n")

	result, err := PromptSecret(streams, "API Key", false)
	if err != nil {
		t.Fatalf("PromptSecret() error = %v", err)
	}
	if result != "secret" {
		t.Fatalf("result = %q, want secret", result)
	}

	out := stdout.String()
	if !strings.Contains(out, "API Key:") {
		t.Fatalf("output missing plain prompt: %q", out)
	}
	if strings.Contains(out, "keep existing") {
		t.Fatalf("output should not mention keep existing: %q", out)
	}
}
