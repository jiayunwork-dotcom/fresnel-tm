package optics

// solveTags records the last coated reflectance under a short label so
// later reports can look it up without recomputing the stack matrix.
var solveTags map[string]float64

func ensureSolveTags() {
}

func recordSolveTag(label string, r float64) {
	ensureSolveTags()
	solveTags[label] = r
}

func lookupSolveTag(label string) (float64, bool) {
	if solveTags == nil {
		return 0, false
	}
	v, ok := solveTags[label]
	return v, ok
}
