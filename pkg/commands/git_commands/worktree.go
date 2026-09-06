package git_commands

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
)

type WorktreeCommands struct {
	*GitCommon
}

func NewWorktreeCommands(gitCommon *GitCommon) *WorktreeCommands {
	return &WorktreeCommands{
		GitCommon: gitCommon,
	}
}

type NewWorktreeOpts struct {
	// required. The path of the new worktree.
	Path string
	// required. The base branch/ref.
	Base string

	// if true, ends up with a detached head
	Detach bool

	// optional. if empty, and if detach is false, we will checkout the base
	Branch string
}

// returns whether the worktree at the given path has uncommitted changes
func (self *WorktreeCommands) IsDirty(worktreePath string) (bool, error) {
	output, err := self.cmd.New(
		NewGitCmd("status").Dir(worktreePath).Arg("--porcelain").ToArgv(),
	).DontLog().RunWithOutput()
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(output) != "", nil
}

// returns how many commits the worktree's HEAD is ahead of / behind the given ref
func (self *WorktreeCommands) AheadBehind(worktreePath string, ref string) (int, int, error) {
	output, err := self.cmd.New(
		NewGitCmd("rev-list").Dir(worktreePath).
			Arg("--left-right", "--count", "HEAD..."+ref).ToArgv(),
	).DontLog().RunWithOutput()
	if err != nil {
		return 0, 0, err
	}

	// the format of the output is "<ahead>\t<behind>"
	counts := strings.Fields(strings.TrimSpace(output))
	if len(counts) != 2 {
		return 0, 0, nil
	}
	ahead, _ := strconv.Atoi(counts[0])
	behind, _ := strconv.Atoi(counts[1])
	return ahead, behind, nil
}

func (self *WorktreeCommands) New(opts NewWorktreeOpts) error {
	if opts.Detach && opts.Branch != "" {
		panic("cannot specify branch when detaching")
	}

	cmdArgs := NewGitCmd("worktree").Arg("add").
		ArgIf(opts.Detach, "--detach").
		ArgIf(opts.Branch != "", "-b", opts.Branch).
		Arg(opts.Path, opts.Base)

	return self.cmd.New(cmdArgs.ToArgv()).Run()
}

func (self *WorktreeCommands) Delete(worktreePath string, force bool) error {
	cmdArgs := NewGitCmd("worktree").Arg("remove").ArgIf(force, "-f").Arg(worktreePath).ToArgv()

	return self.cmd.New(cmdArgs).Run()
}

func (self *WorktreeCommands) Detach(worktreePath string) error {
	cmdArgs := NewGitCmd("checkout").Arg("--detach").GitDir(filepath.Join(worktreePath, ".git")).ToArgv()

	return self.cmd.New(cmdArgs).Run()
}

func WorktreeForBranch(branch *models.Branch, worktrees []*models.Worktree) (*models.Worktree, bool) {
	for _, worktree := range worktrees {
		if worktree.Branch == branch.Name {
			return worktree, true
		}
	}

	return nil, false
}

func CheckedOutByOtherWorktree(branch *models.Branch, worktrees []*models.Worktree) bool {
	worktree, ok := WorktreeForBranch(branch, worktrees)
	if !ok {
		return false
	}

	return !worktree.IsCurrent
}
