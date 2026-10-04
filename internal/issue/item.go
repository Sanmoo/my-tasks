package issue

// Item is one Issue together with the ID that identifies it inside a
// Vault. The ID is the issue file's name — the authority — not a field
// of the frontmatter, so it travels beside the parsed Issue instead of
// inside it. It is the single line through which list, check, show and
// the CLI speak about an Issue.
type Item struct {
	ID    string
	Issue Issue
}
