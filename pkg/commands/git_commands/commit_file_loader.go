package git_commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/commands/oscommands"
	"github.com/jesseduffield/lazygit/pkg/common"
)

type CommitFileLoader struct {
	*common.Common
	cmd oscommands.ICmdObjBuilder
}

func NewCommitFileLoader(common *common.Common, cmd oscommands.ICmdObjBuilder) *CommitFileLoader {
	return &CommitFileLoader{
		Common: common,
		cmd:    cmd,
	}
}

// GetFilesInDiff get the specified commit files
func (self *CommitFileLoader) GetFilesInDiff(from string, to string, reverse bool) ([]*models.CommitFile, error) {
	cmdArgs := NewGitCmd("diff").
		Config("diff.noprefix=false").
		Arg("--submodule").
		Arg("--no-ext-diff").
		Arg("--name-status").
		Arg("-z").
		Arg(fmt.Sprintf("--find-renames=%d%%", self.UserConfig().Git.RenameSimilarityThreshold)).
		ArgIf(reverse, "-R").
		Arg(from).
		Arg(to).
		ToArgv()

	filenames, err := self.cmd.New(cmdArgs).DontLog().RunWithOutput()
	if err != nil {
		return nil, err
	}

	files := getCommitFilesFromFilenames(filenames)
	self.addNumstats(files, func(cmd *GitCommandBuilder) *GitCommandBuilder {
		return cmd.ArgIf(reverse, "-R").Arg(from).Arg(to)
	})
	return files, nil
}

// GetFilesInWorktreeDiff returns the files that differ between the working tree
// (including uncommitted changes) and the given ref
func (self *CommitFileLoader) GetFilesInWorktreeDiff(ref string) ([]*models.CommitFile, error) {
	cmdArgs := NewGitCmd("diff").
		Config("diff.noprefix=false").
		Arg("--submodule").
		Arg("--no-ext-diff").
		Arg("--name-status").
		Arg("-z").
		Arg(fmt.Sprintf("--find-renames=%d%%", self.UserConfig().Git.RenameSimilarityThreshold)).
		Arg(ref).
		ToArgv()

	filenames, err := self.cmd.New(cmdArgs).DontLog().RunWithOutput()
	if err != nil {
		return nil, err
	}

	files := getCommitFilesFromFilenames(filenames)
	self.addNumstats(files, func(cmd *GitCommandBuilder) *GitCommandBuilder {
		return cmd.Arg(ref)
	})
	return files, nil
}

// addNumstats runs the same diff with --numstat and fills in the line change
// counts, keyed by (new) path. Renames in -z numstat output appear as an empty
// third field followed by the two paths as separate entries. Binary files
// ("-" counts) are left at zero. Only runs when numstat display is enabled.
func (self *CommitFileLoader) addNumstats(files []*models.CommitFile, addRefArgs func(*GitCommandBuilder) *GitCommandBuilder) {
	if !self.UserConfig().Gui.ShowNumstatInFilesView || len(files) == 0 {
		return
	}

	cmdArgs := addRefArgs(NewGitCmd("diff").
		Config("diff.noprefix=false").
		Arg("--no-ext-diff").
		Arg("--numstat").
		Arg("-z").
		Arg(fmt.Sprintf("--find-renames=%d%%", self.UserConfig().Git.RenameSimilarityThreshold))).
		ToArgv()

	output, err := self.cmd.New(cmdArgs).DontLog().RunWithOutput()
	if err != nil {
		self.Log.Error(err)
		return
	}

	counts := map[string][2]int{}
	chunks := strings.Split(strings.TrimRight(output, "\x00"), "\x00")
	i := 0
	for i < len(chunks) {
		parts := strings.Split(chunks[i], "	")
		if len(parts) != 3 {
			i++
			continue
		}
		name := parts[2]
		if name == "" && i+2 < len(chunks) {
			// rename entry: the two following chunks are the old and new path
			name = chunks[i+2]
			i += 3
		} else {
			i++
		}
		added, err1 := strconv.Atoi(parts[0])
		deleted, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			continue
		}
		counts[name] = [2]int{added, deleted}
	}

	for _, file := range files {
		if c, ok := counts[file.Path]; ok {
			file.LinesAdded = c[0]
			file.LinesDeleted = c[1]
		}
	}
}

// filenames string is something like "MM\x00file1\x00MU\x00file2\x00AA\x00file3\x00"
// so we need to split it by the null character and then map each status-name pair
// to a commit file. Renames (and copies) are special: their status is followed by
// two paths (the old one and the new one) rather than one, e.g.
// "R100\x00old\x00new\x00".
func getCommitFilesFromFilenames(filenames string) []*models.CommitFile {
	fields := strings.Split(strings.TrimRight(filenames, "\x00"), "\x00")
	if len(fields) == 1 {
		return []*models.CommitFile{}
	}

	commitFiles := make([]*models.CommitFile, 0, len(fields)/2)
	for i := 0; i < len(fields)-1; {
		changeStatus := fields[i]
		if changeStatus[0] == 'R' || changeStatus[0] == 'C' {
			// The status has a similarity score appended (e.g. "R100"); drop it
			// so the rest of the code only has to deal with a plain "R" or "C".
			commitFiles = append(commitFiles, &models.CommitFile{
				ChangeStatus: changeStatus[:1],
				PreviousPath: fields[i+1],
				Path:         fields[i+2],
			})
			i += 3
		} else {
			// typical result looks like 'A my_file' meaning my_file was added
			commitFiles = append(commitFiles, &models.CommitFile{
				ChangeStatus: changeStatus,
				Path:         fields[i+1],
			})
			i += 2
		}
	}

	return commitFiles
}
