package context

import (
	"fmt"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/gui/filetree"
	"github.com/jesseduffield/lazygit/pkg/gui/presentation"
	"github.com/jesseduffield/lazygit/pkg/gui/presentation/icons"
	"github.com/jesseduffield/lazygit/pkg/gui/style"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/jesseduffield/lazygit/pkg/utils"
	"github.com/samber/lo"
)

// FilesFromMainContext shows the files that differ between the working tree and
// the merge base with the main branch (like a pull request's files view).
type FilesFromMainContext struct {
	*filetree.CommitFileTreeViewModel
	*ListContextTrait

	// the merge base with the main branch that the file list was last computed
	// against; set alongside the model write, read to build the diff command
	baseRef string
}

var (
	_ types.IListContext       = (*FilesFromMainContext)(nil)
	_ types.IFilterableContext = (*FilesFromMainContext)(nil)
)

func NewFilesFromMainContext(c *ContextCommon) *FilesFromMainContext {
	viewModel := filetree.NewCommitFileTreeViewModel(
		func() []*models.CommitFile { return c.Model().FilesFromMain },
		c.Common,
		c.UserConfig().Gui.ShowFileTree,
	)
	// the tree renderer resolves patch status against the model's ref, so it
	// must never be nil, even before the first refresh has computed the base
	viewModel.SetRef(&mergeBaseRef{hash: ""})

	getDisplayStrings := func(_ int, _ int) [][]string {
		// while another side panel is active, show a one-line summary instead of
		// the top of the file list. Keying off the current side context (rather
		// than the current context) keeps the file list showing while a
		// non-side context above this panel has focus, e.g. the search prompt or
		// the main view.
		if current := c.Context().CurrentSide(); current == nil || current.GetKey() != FILES_FROM_MAIN_CONTEXT_KEY {
			count := len(c.Model().FilesFromMain)
			return [][]string{{style.FgYellow.Sprint(utils.ResolvePlaceholderString(
				c.Tr.FilesChangedFromMain,
				map[string]string{"count": fmt.Sprint(count)},
			))}}
		}

		if viewModel.Len() == 0 {
			return [][]string{{style.FgRed.Sprint("(none)")}}
		}

		showFileIcons := icons.IsIconEnabled() && c.UserConfig().Gui.ShowFileIcons
		lines := presentation.RenderCommitFileTree(viewModel, c.Git().Patch.PatchBuilder, showFileIcons, &c.UserConfig().Gui.CustomIcons)
		return lo.Map(lines, func(line string, _ int) []string {
			return []string{line}
		})
	}

	ctx := &FilesFromMainContext{
		CommitFileTreeViewModel: viewModel,
		ListContextTrait: &ListContextTrait{
			Context: NewSimpleContext(
				NewBaseContext(NewBaseContextOpts{
					View:       c.Views().FilesFromMain,
					WindowName: "files",
					Key:        FILES_FROM_MAIN_CONTEXT_KEY,
					Kind:       types.SIDE_CONTEXT,
					Focusable:  true,
				}),
			),
			ListRenderer: ListRenderer{
				list:              viewModel,
				getDisplayStrings: getDisplayStrings,
			},
			c: c,
		},
	}

	return ctx
}

func (self *FilesFromMainContext) GetBaseRef() string {
	return self.baseRef
}

func (self *FilesFromMainContext) SetBaseRef(baseRef string) {
	self.baseRef = baseRef
	self.SetRef(&mergeBaseRef{hash: baseRef})
}

// mergeBaseRef is a minimal models.Ref for the merge base commit, needed by the
// commit file tree's rendering (which keys the patch builder off a ref name)
type mergeBaseRef struct {
	hash string
}

var _ models.Ref = (*mergeBaseRef)(nil)

func (self *mergeBaseRef) FullRefName() string   { return self.hash }
func (self *mergeBaseRef) RefName() string       { return self.hash }
func (self *mergeBaseRef) ShortRefName() string  { return utils.ShortHash(self.hash) }
func (self *mergeBaseRef) ParentRefName() string { return self.hash + "^" }
func (self *mergeBaseRef) Description() string   { return "merge base with main" }
