package rules

// Movement beyond walk and run at 1.65, from the rules-movement capability:
// sprint, swimming, and the jump and gravity the shared step uses.

const (
	// SprintFactor multiplies run speed while sprinting.
	SprintFactor = 1.25
	// SprintCost is the endurance a second of sprinting while moving takes.
	SprintCost = 5.0
	// SwimSpeed is the speed of a swimming body, in units a second.
	SwimSpeed = 120.0
	// JumpSpeed is the upward speed a jump starts with, in units a second.
	JumpSpeed = 210.0
	// Gravity is the downward pull on a body in the air, units a second squared.
	Gravity = 500.0
	// StepDown is the most the ground may drop in one step before the body
	// leaves it and falls rather than following it down.
	StepDown = 12.0
	// SwimDepth is how deep the water must be over the ground for a body to
	// swim in it, in units.
	SwimDepth = 50.0
)

// SprintSpeed is the speed of a sprint from the run speed under the same
// effects.
func SprintSpeed(run float64) float64 { return run * SprintFactor }

// SprintContinues says whether a sprinter with this much endurance may go on.
func SprintContinues(endurance float64) bool { return endurance > 0 }

// JumpAirSteps is how many 1/60 s steps a jump on flat ground stays in the
// air, integrating as the movement step does: speed first, then height.
func JumpAirSteps() int {
	const dt = 1.0 / 60
	z, vz, n := 0.0, JumpSpeed, 0
	for {
		vz -= Gravity * dt
		z += vz * dt
		n++
		if z <= 0 || n > 10000 {
			return n
		}
	}
}

func init() {
	const c = "rules-movement"
	register(
		Rule{ID: "movement.sprint", Capability: c, Requirement: "Sprint speed and cost", Status: Provisional,
			Sources:   []string{"project"},
			Scenarios: []string{"Sprint speed", "Standing sprint costs nothing", "Out of endurance"}},
		Rule{ID: "movement.swim", Capability: c, Requirement: "Swim speed", Status: Provisional,
			Sources: []string{"project"}, Scenarios: []string{"Swimming"}},
		Rule{ID: "movement.jump", Capability: c, Requirement: "Jump and gravity", Status: Provisional,
			Sources: []string{"project"}, Scenarios: []string{"Airtime"}},
	)
}

// SprintDrain is the endurance a stretch of sprinting costs: only the time
// spent moving counts.
func SprintDrain(seconds float64, moving bool) float64 {
	if !moving {
		return 0
	}
	return seconds * SprintCost
}
