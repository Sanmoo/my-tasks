package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-tasks2/internal/check"
	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/priority"
	"github.com/Sanmoo/my-tasks2/internal/vault"
)

// newCheckCmd builds `mt check`: audits the vault's Issue files and reports
// Rank, frontmatter, Status and datetime findings. --fix repairs only the
// ranked queue; other findings remain errors to be corrected by the user.
func newCheckCmd() *cobra.Command {
	var fix bool
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Audit a Vault's Issues",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return exitcode.Usage(fmt.Errorf("check takes no arguments"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(cmd, fix)
		},
	}
	cmd.Flags().BoolVar(&fix, "fix", false, "renormalize Ranks to 1..N")
	return cmd
}

// runCheck audits a vault's Issues. Duplicate Ranks are errors because they
// make the queue ambiguous; Rank gaps are warnings because they do not make
// the queue ambiguous and can be repaired by --fix.
func runCheck(cmd *cobra.Command, fix bool) error {
	vaultDir, err := resolveVault(cmd)
	if err != nil {
		return err
	}
	vcfg, err := vault.LoadVault(vaultDir)
	if err != nil {
		return err
	}
	items, err := loadCheckItems(vaultDir)
	if err != nil {
		return err
	}
	if err := validateVault(vcfg, items); err != nil {
		return err
	}
	if fix {
		changes := priority.RenormalizeRanks(priorityIssuesFromCheckItems(items))
		for _, change := range changes {
			if err := applyCheckRankChange(vaultDir, change); err != nil {
				return err
			}
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Fixed %d Ranks\n", len(changes))
		items, err = loadCheckItems(vaultDir)
		if err != nil {
			return err
		}
		if err := validateVault(vcfg, items); err != nil {
			return err
		}
	}
	if invalid := check.NonPositiveRanks(items); len(invalid) > 0 {
		return fmt.Errorf("invalid rank: %s (Ranks must be greater than zero)", formatRanks(invalid))
	}
	if gaps := check.RankGapRanges(items); len(gaps) > 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: rank gap: %s\n", formatRankGaps(gaps))
	}
	if dups := check.DuplicateRanks(items); len(dups) > 0 {
		return fmt.Errorf("duplicate rank: %s", duplicateRankDetails(items, dups))
	}
	fmt.Fprintln(cmd.OutOrStdout(), "OK")
	return nil
}

func loadCheckItems(vaultDir string) ([]check.Item, error) {
	files, err := readIssueFiles(vaultDir)
	if err != nil {
		return nil, fmt.Errorf("malformed frontmatter: %w", err)
	}
	items := make([]check.Item, 0, len(files))
	for _, file := range files {
		if err := check.ValidateFrontmatter(file.Data, file.ID); err != nil {
			return nil, err
		}
		items = append(items, check.Item{ID: file.ID, Issue: file.Issue})
	}
	return items, nil
}

func duplicateRankDetails(items []check.Item, ranks []int) string {
	byRank := make(map[int][]string)
	for _, item := range items {
		if item.Issue.Frontmatter.Rank != nil {
			byRank[*item.Issue.Frontmatter.Rank] = append(byRank[*item.Issue.Frontmatter.Rank], item.ID)
		}
	}
	details := make([]string, len(ranks))
	for i, rank := range ranks {
		details[i] = fmt.Sprintf("%d (%s)", rank, strings.Join(byRank[rank], ", "))
	}
	return strings.Join(details, "; ")
}

func priorityIssuesFromCheckItems(items []check.Item) []priority.Issue {
	issues := make([]priority.Issue, 0, len(items))
	for _, item := range items {
		issues = append(issues, priorityIssueFrom(item.ID, item.Issue))
	}
	return issues
}

func applyCheckRankChange(vaultDir string, change priority.Change) error {
	i, err := readIssue(vaultDir, change.ID)
	if err != nil {
		return err
	}
	i.Frontmatter.Rank = change.Rank
	return writeIssueFile(vaultDir, change.ID, i)
}

// validateVault runs the per-Issue schema checks (frontmatter values,
// status, datetime layout) and the vault-wide blocked_by reference
// checks (existence, self-block, cycles). It returns the first
// violation found.
func validateVault(vcfg vault.Vault, items []check.Item) error {
	statuses := vcfg.StatusList()
	for _, item := range items {
		if err := check.ValidateItem(item, statuses); err != nil {
			return err
		}
	}
	return check.ValidateBlockedBy(items)
}

func formatRanks(ranks []int) string {
	values := make([]string, len(ranks))
	for i, rank := range ranks {
		values[i] = strconv.Itoa(rank)
	}
	return strings.Join(values, ", ")
}

func formatRankGaps(gaps []check.RankGap) string {
	values := make([]string, len(gaps))
	for i, gap := range gaps {
		if gap.Start == gap.End {
			values[i] = strconv.Itoa(gap.Start)
			continue
		}
		values[i] = fmt.Sprintf("%d-%d", gap.Start, gap.End)
	}
	return strings.Join(values, ", ")
}
