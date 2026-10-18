package issuefiles

// faultStage names a point of ApplyRankPlan where a test may force a
// failure. The hook is nil in production and set only from export_test.go.
type faultStage int

const (
	stageBeforeWrite  faultStage = iota // before an Issue is written
	stageAfterWrite                     // after an Issue was written (the file may hold the new bytes)
	stageCreated                        // after the created Issue was written
	stageKeep                           // while the originals are copied to the recovery directory
	stageBeforeCreate                   // before the created Issue's open (the hook may plant a file at the name)
	stageRestore                        // before an Issue is restored
	stageCleanup                        // before the recovery directory is removed
)

// inject returns the error a test hook forces at stage for id, or nil.
var faultHook func(stage faultStage, id string) error

func inject(stage faultStage, id string) error {
	if faultHook == nil {
		return nil
	}
	return faultHook(stage, id)
}
