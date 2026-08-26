package recipe

import (
	"os"
	"path/filepath"
	"testing"

	"fresnel-tm/internal/model"
	"fresnel-tm/internal/stackgen"
)

func TestRecipeRoundTripSolveMatches(t *testing.T) {
	s, err := stackgen.SingleAR(1.38, 550)
	if err != nil {
		t.Fatal(err)
	}
	rec, err := Seal(s, model.Incidence{WavelengthNm: 550, AngleDeg: 0}, model.PolAverage)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Reflection >= rec.Transmission && rec.Reflection > 0.05 {
		t.Fatalf("AR recipe R looks too high: %g", rec.Reflection)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "coat.frtm")
	if err := Create(path, rec); err != nil {
		t.Fatal(err)
	}
	got, err := ReplayAndVerify(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("records %d", len(got))
	}
	if got[0].Reflection != rec.Reflection {
		t.Fatalf("replay R %g != sealed %g", got[0].Reflection, rec.Reflection)
	}
}

func TestTruncatedTailDoesNotPollutePrefix(t *testing.T) {
	ar, err := stackgen.SingleAR(1.38, 550)
	if err != nil {
		t.Fatal(err)
	}
	first, err := Seal(ar, model.Incidence{WavelengthNm: 550, AngleDeg: 0}, model.PolAverage)
	if err != nil {
		t.Fatal(err)
	}
	bragg, err := stackgen.Bragg(2.3, 1.38, 3, 550)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Seal(bragg, model.Incidence{WavelengthNm: 550, AngleDeg: 0}, model.PolS)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.frtm")
	if err := Create(path, first); err != nil {
		t.Fatal(err)
	}
	if err := Commit(path, second); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if st.Size() < 80 {
		t.Fatalf("journal too small: %d", st.Size())
	}
	if err := os.Truncate(path, st.Size()-17); err != nil {
		t.Fatal(err)
	}
	got, err := ReplayAndVerify(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("truncated tail should leave only the committed prefix, got %d", len(got))
	}
	if got[0].Reflection != first.Reflection {
		t.Fatalf("prefix R %g != first %g", got[0].Reflection, first.Reflection)
	}
	if err := Verify(got[0]); err != nil {
		t.Fatal(err)
	}
}
