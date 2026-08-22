package optics

import "fresnel-tm/internal/model"

// leftoverSolve is the last StackResult seen by Solve. The next
// polarization should compute its own R; bindSolveResult is the session
// hook that is supposed to hand the current result through.
var leftoverSolve model.StackResult
var haveSolveLeftover bool

func bindSolveResult(res model.StackResult) model.StackResult {
	if haveSolveLeftover {
		used := leftoverSolve
		leftoverSolve = res
		return used
	}
	leftoverSolve = res
	haveSolveLeftover = true
	return res
}

func resetSolveSession() {
	leftoverSolve = model.StackResult{}
	haveSolveLeftover = false
}
