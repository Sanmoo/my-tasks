package issuefiles

// FaultStage and its constants expose the package's fault seam to the
// external tests. The seam exists only in test builds.
type FaultStage = faultStage

const (
	StageBeforeWrite  = stageBeforeWrite
	StageAfterWrite   = stageAfterWrite
	StageKeep         = stageKeep
	StageBeforeCreate = stageBeforeCreate
	StageCreated      = stageCreated
	StageRestore      = stageRestore
	StageCleanup      = stageCleanup
)

// SetFaultHook installs hook as the fault seam and returns the function
// that removes it. Tests using it must not run in parallel.
func SetFaultHook(hook func(stage FaultStage, id string) error) func() {
	faultHook = hook
	return func() { faultHook = nil }
}
