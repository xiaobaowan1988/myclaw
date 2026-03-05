package ui

import (
	"strings"
	"testing"
)

func TestPrinterColor(t *testing.T) {
	p := NewPrinter(false)
	result := p.color(Red, "test")
	if !strings.Contains(result, "\033[31m") {
		t.Error("expected ANSI color code")
	}
	if !strings.Contains(result, "test") {
		t.Error("expected text content")
	}

	// No color mode
	p2 := NewPrinter(true)
	result2 := p2.color(Red, "test")
	if strings.Contains(result2, "\033[") {
		t.Error("should not contain ANSI codes in no-color mode")
	}
	if result2 != "test" {
		t.Errorf("expected 'test', got '%s'", result2)
	}
}

func TestSpinnerStartStop(t *testing.T) {
	s := NewSpinner("loading")
	s.Start()
	s.Stop()
	// Should not panic on double stop
	s.Stop()
}
