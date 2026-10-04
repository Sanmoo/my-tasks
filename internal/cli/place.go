// Package cli — the mt place command and the Encaixe session it shares with
// create/q. The session itself (the binary search over the queue) lives in
// internal/priority; here live the process concerns: the terminal gate, the
// answers, the Comparação rendering and applying the resulting rank.
package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/issue"
	"github.com/Sanmoo/my-tasks2/internal/list"
	"github.com/Sanmoo/my-tasks2/internal/priority"
)

// errPlaceAborted cancels an Encaixe session: nothing was written. The
// commands decide what a cancellation means — `mt place` fails, because the
// session was its whole job, while create/q keep the Issue they captured in
// the Backlog.
var errPlaceAborted = errors.New("placement cancelled")

// answersUsage documents the --answers sequence shared by place, create and q.
const answersUsage = "answers for a non-interactive session, in order: a (the Issue first), b (the Candidate first), i (indifferent); q cancels"

// placeRequest is the Encaixe session a create/q run carries: answers feeds it
// non-interactively, disabled suppresses it so the new Issue stays in the
// Backlog — exactly what create did before the session existed.
type placeRequest struct {
	answers  string
	disabled bool
}

// newPlaceCmd builds `mt place <id>`: it asks which of two Issues is more
// prioritized until the rank of id is known — the optimal ⌈log2(N+1)⌉
// Comparações for a queue of N — then applies it.
func newPlaceCmd() *cobra.Command {
	var answers string
	cmd := &cobra.Command{
		Use:   "place <id>",
		Short: "Place an Issue in the queue by pairwise comparison",
		Long:  placeLong,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return exitcode.Usage(fmt.Errorf("place needs exactly one issue ID"))
			}
			return nil
		},
		ValidArgsFunction: completeIssueID,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPlace(cmd, args[0], answers)
		},
	}
	cmd.Flags().StringVar(&answers, "answers", "", answersUsage)
	return cmd
}

// runPlace is the explicit Encaixe flow: validate the key and the vault's
// queue, then discover the rank and print it. Every precondition fails before
// the first Comparação, and without a terminal (or --answers) there is no
// session to run — which is an error here, not a silent no-op.
func runPlace(cmd *cobra.Command, id, answers string) error {
	t, err := resolveVaultForKey(cmd, id)
	if err != nil {
		return err
	}
	items, err := loadItems(t.dir)
	if err != nil {
		return err
	}
	if (t.byPrefix != "" || t.hint != "") && !containsItem(items, id) {
		return t.notFoundError(id)
	}
	if _, err := priority.PlacementTarget(items, id); err != nil {
		return err
	}
	if err := checkNoDuplicateRanks(items); err != nil {
		return err
	}
	if !sessionAvailable(cmd, answers) {
		return fmt.Errorf("place needs a terminal to ask the Comparações; use --answers for a non-interactive session")
	}
	rank, err := runPlaceSession(cmd, t.dir, items, id, answers)
	if errors.Is(err, errPlaceAborted) {
		return fmt.Errorf("placement of %s cancelled: the queue is unchanged", id)
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Placed %s at rank %d\n", id, rank)
	return nil
}

// placeNewIssue runs the implicit session for a just-created Issue. A session
// that cannot run is not an error: create's job was to capture the Issue, and
// the Backlog is where an unprioritized Issue belongs. A duplicate rank is
// reported — it would make the session's promise ambiguous — and no terminal
// simply means no human to ask.
func placeNewIssue(cmd *cobra.Command, vaultDir, id, answers string) error {
	out := cmd.ErrOrStderr()
	items, err := loadItems(vaultDir)
	if err != nil {
		return err
	}
	if err := checkNoDuplicateRanks(items); err != nil {
		fmt.Fprintf(out, "Warning: %s, so %s stays in the Backlog\n", err, id)
		return nil
	}
	if !sessionAvailable(cmd, answers) {
		return nil
	}
	rank, err := runPlaceSession(cmd, vaultDir, items, id, answers)
	switch {
	case errors.Is(err, errPlaceAborted):
		// The Issue is already saved; the session was the optional part. The
		// explicit `mt place` is how the user comes back to it.
		fmt.Fprintf(out, "%s stays in the Backlog (mt place %s reopens the session)\n", id, id)
		return nil
	case err != nil:
		return err
	}
	fmt.Fprintf(out, "Placed %s at rank %d\n", id, rank)
	return nil
}

// runPlaceSession asks the Comparações for id and returns the rank they
// converge to, applying it through the quick-order plan (the single source of
// truth for the resulting rank shifts). items must be the vault as it is on
// disk, id included. Nothing is written until the session ends, so a cancelled
// session leaves the queue untouched.
func runPlaceSession(cmd *cobra.Command, vaultDir string, items []issue.Item, id, answers string) (int, error) {
	target, err := priority.PlacementTarget(items, id)
	if err != nil {
		return 0, err
	}
	out := cmd.ErrOrStderr()
	place := priority.NewPlace(priority.Candidates(items, id))
	replies := newAnswerReader(cmd, answers)
	for {
		candidate, ok := place.Question()
		if !ok {
			break
		}
		fmt.Fprintf(out, "Which is more prioritized? (%d/%d)\n", place.Asked()+1, place.Total())
		fmt.Fprintf(out, "  a) %s\n", placeLine(target))
		fmt.Fprintf(out, "  b) %s\n", placeLine(candidate))
		fmt.Fprint(out, "a (first), b (second) or i (indifferent); q cancels: ")
		answer, ok, err := replies.next()
		if err != nil {
			return 0, err
		}
		if !ok {
			return 0, errPlaceAborted
		}
		place.Answer(answer)
	}
	rank := place.Rank()
	changes, err := priority.QuickPlan(items, id, priority.MoveToRank, rank)
	if err != nil {
		return 0, err
	}
	if err := applyRankChanges(vaultDir, changes); err != nil {
		return 0, err
	}
	return rank, nil
}

// sessionAvailable reports whether a Comparação session can run: a batch
// --answers sequence makes it non-interactive, otherwise it needs a human at a
// terminal. The gate is stdout, not stdin — in `ID=$(mt q "x")` stdin is still
// the terminal, and a session there would ask into the captured output.
func sessionAvailable(cmd *cobra.Command, answers string) bool {
	return answers != "" || stdoutIsTerminal(cmd.OutOrStdout())
}

// stdoutIsTerminal reports whether w is an interactive terminal. Only the real
// process streams are *os.File; an in-process caller of Run passes a writer,
// which is then never a terminal. It is the same rule show uses for color.
func stdoutIsTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

// checkNoDuplicateRanks refuses the Encaixe when the vault holds a duplicate
// rank: the queue order is ambiguous for pick-next, so a session could not
// promise a position the rest of the tool honours. Renormalizing is
// `mt check --fix`'s job, not a silent side effect of asking questions.
func checkNoDuplicateRanks(items []issue.Item) error {
	if dups := list.DuplicateRanks(items); len(dups) > 0 {
		return fmt.Errorf("duplicate rank: %d (run mt check --fix)", dups[0])
	}
	return nil
}

// placeLine renders one side of a Comparação: the compact one-line view of
// `mt list` — the domain's Item already carries everything the line needs.
func placeLine(item issue.Item) string {
	return list.FormatLine(item)
}

// answerReader yields the replies to a session's Comparações: from the batch
// --answers sequence when one was given, else one line at a time from stdin.
type answerReader struct {
	out   io.Writer
	stdin *bufio.Reader
	// batch holds the remaining --answers runes; batchOn records that a
	// sequence was given at all, so an exhausted one is an error instead of a
	// fallback to stdin.
	batch   []rune
	batchOn bool
}

// newAnswerReader builds the reader of one session. An empty answers string is
// not a batch: the session reads stdin.
func newAnswerReader(cmd *cobra.Command, answers string) *answerReader {
	return &answerReader{
		out:     cmd.ErrOrStderr(),
		stdin:   bufio.NewReader(cmd.InOrStdin()),
		batch:   []rune(answers),
		batchOn: answers != "",
	}
}

// next returns the next Answer. ok is false when the session was cancelled —
// the q key or the end of input. err is a usage error for a batch that is not
// a valid answer sequence.
func (r *answerReader) next() (priority.Answer, bool, error) {
	for {
		key, err := r.key()
		if errors.Is(err, io.EOF) {
			return 0, false, nil
		}
		if err != nil {
			return 0, false, err
		}
		switch key {
		case 'a':
			return priority.TargetBefore, true, nil
		case 'b':
			return priority.TargetAfter, true, nil
		case 'i':
			return priority.Indifferent, true, nil
		case 'q':
			return 0, false, nil
		}
		if r.batchOn {
			return 0, false, exitcode.Usage(fmt.Errorf("--answers takes a, b, i or q, not %q", string(key)))
		}
		// Interactive: a key that is not an answer re-asks instead of failing
		// the whole session.
		fmt.Fprintln(r.out, "Answer a, b, i or q.")
	}
}

// key returns the next key of the session: the next rune of the batch, or the
// next line of stdin. A line that is not a single key yields 0, which is never
// a valid answer, so next re-asks (or rejects it, in batch mode).
func (r *answerReader) key() (rune, error) {
	if r.batchOn {
		if len(r.batch) == 0 {
			return 0, exitcode.Usage(errors.New("--answers ran out before the session ended"))
		}
		key := r.batch[0]
		r.batch = r.batch[1:]
		return unicode.ToLower(key), nil
	}
	for {
		line, err := r.stdin.ReadString('\n')
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			runes := []rune(trimmed)
			if len(runes) != 1 {
				return 0, nil
			}
			return unicode.ToLower(runes[0]), nil
		}
		if err != nil {
			return 0, err
		}
	}
}

const placeLong = `place discovers where an Issue belongs in the priority queue by asking, one
Comparação at a time, which of two Issues is more prioritized:

  Which is more prioritized? (1/5)
    a) ○ pkm-055  comprar material
    b) ○ pkm-07r0  revisar orçamento
  a (first), b (second) or i (indifferent); q cancels:

Each answer halves the interval of possible ranks, so a queue of N Issues asks
at most ⌈log2(N+1)⌉ Comparações — the fewest an insertion into an ordered queue
can take. Option a is always the Issue being placed. The reference is the queue
itself: ranked open and in_progress Issues. Backlog and finished Issues are
never candidates. Answering i keeps the order the vault already has: the placed
Issue goes right after the Candidate of that Comparação.

Nothing is written until the session ends; q, Ctrl-C or the end of input
cancels it and leaves the queue untouched. place needs a terminal to ask in, or
--answers for a scripted session. A vault with a duplicate rank is refused
before the first Comparação — run mt check --fix first.`
