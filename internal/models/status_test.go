package models

import (
	"testing"
	"time"
)

func TestComputeContainerItemStatus(t *testing.T) {
	now := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	at := func(d time.Duration) time.Time { return now.Add(d) }

	cases := []struct {
		name      string
		hasBL     bool
		etd, eta  time.Time
		loadQty   float64
		landedQty float64
		want      string
	}{
		{"1. no BL → Expected", false, time.Time{}, time.Time{}, 100, 0, "Expected"},
		{"2. before ETD → BL Released", true, at(24 * time.Hour), at(10 * 24 * time.Hour), 100, 0, "BL Released"},
		{"3. ETD~ETA → Shipping", true, at(-24 * time.Hour), at(24 * time.Hour), 100, 0, "Shipping"},
		{"4. after ETA, no GR → Arrived", true, at(-10 * 24 * time.Hour), at(-24 * time.Hour), 100, 0, "Arrived"},
		{"5. partial GR → Partially Landed", true, at(-10 * 24 * time.Hour), at(-24 * time.Hour), 100, 40, "Partially Landed"},
		{"6. full GR → Landed", true, at(-10 * 24 * time.Hour), at(-24 * time.Hour), 100, 100, "Landed"},
		{"over GR → Landed", true, at(-10 * 24 * time.Hour), at(-24 * time.Hour), 100, 120, "Landed"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ComputeContainerItemStatus(c.hasBL, c.etd, c.eta, c.loadQty, c.landedQty, now)
			if got != c.want {
				t.Fatalf("want %q, got %q", c.want, got)
			}
		})
	}
}

func TestComputeClearanceStatus(t *testing.T) {
	cases := []struct {
		cleared, total int
		want           string
	}{
		{0, 0, "Not Cleared"},
		{0, 2, "Not Cleared"},
		{1, 2, "Partially Cleared"},
		{2, 2, "All Cleared"},
		{3, 2, "All Cleared"},
	}
	for _, c := range cases {
		got := ComputeClearanceStatus(c.cleared, c.total)
		if got != c.want {
			t.Errorf("cleared=%d total=%d: want %q, got %q", c.cleared, c.total, c.want, got)
		}
	}
}

func TestComputeAPAllocationStatus(t *testing.T) {
	cases := []struct {
		name      string
		local     float64
		allocated float64
		hasAlloc  bool
		current   string
		want      string
	}{
		{"full allocation → Closed", 1000, 1000, true, "Open", "Closed"},
		{"over allocation → Closed", 1000, 1200, true, "Open", "Closed"},
		{"float noise → Closed", 1000, 999.995, true, "Open", "Closed"},
		{"partial allocation → Open", 1000, 400, true, "Draft", "Open"},
		{"no allocation, Draft 유지", 1000, 0, false, "Draft", "Draft"},
		{"no allocation, Open 유지", 1000, 0, false, "Open", "Open"},
		{"no allocation, 이전 Closed → Open", 1000, 0, false, "Closed", "Open"},
		{"credit full → Closed", -500, -500, true, "Open", "Closed"},
		{"credit partial → Open", -500, -200, true, "Open", "Open"},
		{"credit over → Closed", -500, -600, true, "Open", "Closed"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ComputeAPAllocationStatus(c.local, c.allocated, c.hasAlloc, c.current)
			if got != c.want {
				t.Fatalf("want %q, got %q", c.want, got)
			}
		})
	}
}
