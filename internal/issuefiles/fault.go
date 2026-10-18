package issuefiles

// faultStage names a point of ApplyRankPlan where a test may force a
// failure. The hook is the package's only fault seam: it is nil in
// production, so inject is a no-op there, and it is set only from an
// export_test.go file.
type faultStage int

const (
	// stageBeforeWrite fails before an Issue is written: the file is
	// untouched.
	stageBeforeWrite faultStage = iota
	// stageAfterWrite fails after an Issue was written: the file has its
	// new bytes, as after a write that returned an error midway.
	stageAfterWrite
	// stageCreated fails after the created Issue was written.
	stageCreated
	// stageKeep fails while the originals are being copied to the
	// recovery directory: a preparation failure, before any Issue is
	// written.
	stageKeep
	// stageBeforeCreate fails before the created Issue's open: the test
	// hook may put a file at the target name first, so the open fails
	// with EEXIST while the entry was never this operation's.
	stageBeforeCreate
	// stageRestore fails before an Issue is restored.
	stageRestore
	// stageCleanup fails before the recovery directory is removed.
	stageCleanup
)

// faultHook, when set, is consulted at every faultStage.
var faultHook func(stage faultStage, id string) error

// inject returns the error a test hook forces at stage for id, or nil.
func inject(stage faultStage, id string) error {
	if faultHook == nil {
		return nil
	}
	return faultHook(stage, id)
}
