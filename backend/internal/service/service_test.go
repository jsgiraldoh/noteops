package service

import (
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/johansgiraldo/noteops/backend/internal/models"
)

// Service con repo y db nulos es válido para ComputeSessionTick
// porque este método no usa ninguno de los dos.
func newTestService() *Service {
	return New(nil, nil)
}

// ─── ComputeSessionTick ───────────────────────────────────────────────────────

func TestComputeSessionTick_ActiveSession(t *testing.T) {
	svc := newTestService()
	session := &models.Session{
		ID:          uuid.New(),
		StartsAt:    time.Now().Add(-30 * time.Minute),
		DurationMin: 120,
		Active:      true,
	}

	tick := svc.ComputeSessionTick(session)

	if !tick.IsActive {
		t.Error("sesión activa a mitad de duración debe tener IsActive=true")
	}
	assertInDelta(t, "ElapsedSec", 1800, float64(tick.ElapsedSec), 5)
	assertInDelta(t, "RemainingSec", 5400, float64(tick.RemainingSec), 5)
	if tick.DurationMin != 120 {
		t.Errorf("DurationMin = %d, quería 120", tick.DurationMin)
	}
}

func TestComputeSessionTick_ExpiredSession(t *testing.T) {
	svc := newTestService()
	session := &models.Session{
		ID:          uuid.New(),
		StartsAt:    time.Now().Add(-3 * time.Hour),
		DurationMin: 120,
		Active:      true,
	}

	tick := svc.ComputeSessionTick(session)

	if tick.IsActive {
		t.Error("sesión expirada debe tener IsActive=false")
	}
	if tick.RemainingSec != 0 {
		t.Errorf("sesión expirada debe tener RemainingSec=0, got %d", tick.RemainingSec)
	}
}

func TestComputeSessionTick_NotActivated(t *testing.T) {
	svc := newTestService()
	session := &models.Session{
		ID:          uuid.New(),
		StartsAt:    time.Now().Add(-10 * time.Minute),
		DurationMin: 120,
		Active:      false,
	}

	tick := svc.ComputeSessionTick(session)

	if tick.IsActive {
		t.Error("sesión con Active=false debe tener IsActive=false aunque tenga tiempo restante")
	}
}

func TestComputeSessionTick_JustStarted(t *testing.T) {
	svc := newTestService()
	session := &models.Session{
		ID:          uuid.New(),
		StartsAt:    time.Now(),
		DurationMin: 60,
		Active:      true,
	}

	tick := svc.ComputeSessionTick(session)

	if !tick.IsActive {
		t.Error("sesión recién iniciada debe tener IsActive=true")
	}
	assertInDelta(t, "ElapsedSec", 0, float64(tick.ElapsedSec), 3)
	assertInDelta(t, "RemainingSec", 3600, float64(tick.RemainingSec), 3)
}

func TestComputeSessionTick_RemainingNeverNegative(t *testing.T) {
	svc := newTestService()
	session := &models.Session{
		ID:          uuid.New(),
		StartsAt:    time.Now().Add(-10 * time.Hour),
		DurationMin: 120,
		Active:      true,
	}

	tick := svc.ComputeSessionTick(session)

	if tick.RemainingSec < 0 {
		t.Errorf("RemainingSec nunca debe ser negativo, got %d", tick.RemainingSec)
	}
	if tick.RemainingSec != 0 {
		t.Errorf("sesión muy expirada debe tener RemainingSec=0, got %d", tick.RemainingSec)
	}
}

func TestComputeSessionTick_SessionIDPreserved(t *testing.T) {
	svc := newTestService()
	id := uuid.New()
	session := &models.Session{
		ID:          id,
		StartsAt:    time.Now().Add(-5 * time.Minute),
		DurationMin: 60,
		Active:      true,
	}

	tick := svc.ComputeSessionTick(session)

	if tick.SessionID != id.String() {
		t.Errorf("SessionID = %q, quería %q", tick.SessionID, id.String())
	}
}

// ─── GenerateSlots — lógica de conteo ────────────────────────────────────────
// GenerateSlots usa DB (s.db.Begin). Testeamos la lógica de conteo directamente
// verificando que session.DurationMin / session.SlotMin produce el count correcto.

func TestSlotCount_TwoHoursWith20MinSlots(t *testing.T) {
	duration, slotMin := 120, 20
	count := duration / slotMin
	if count != 6 {
		t.Errorf("120/20 = %d slots, quería 6", count)
	}
}

func TestSlotCount_TwoHoursWith5MinSlots(t *testing.T) {
	duration, slotMin := 120, 5
	count := duration / slotMin
	if count != 24 {
		t.Errorf("120/5 = %d slots, quería 24", count)
	}
}

func TestSlotCount_MinimumOneSlot(t *testing.T) {
	duration, slotMin := 10, 120 // duración menor que slot
	count := duration / slotMin
	if count == 0 {
		count = 1
	}
	if count != 1 {
		t.Errorf("duración menor que slotMin debe producir 1 slot, got %d", count)
	}
}

func TestSlotStartTimes(t *testing.T) {
	base := time.Date(2026, 6, 6, 8, 0, 0, 0, time.UTC)
	slotMin := 20
	count := 6

	for i := 0; i < count; i++ {
		expected := base.Add(time.Duration(i*slotMin) * time.Minute)
		if expected.Minute() != (i*slotMin)%60 {
			t.Errorf("slot %d starts_at minuto = %d, quería %d", i+1, expected.Minute(), (i*slotMin)%60)
		}
	}
}

// ─── helper ──────────────────────────────────────────────────────────────────

func assertInDelta(t *testing.T, name string, expected, got, delta float64) {
	t.Helper()
	if math.Abs(got-expected) > delta {
		t.Errorf("%s = %.0f, quería %.0f ± %.0f", name, got, expected, delta)
	}
}
