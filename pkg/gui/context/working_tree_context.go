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

type WorkingTreeContext struct {
	*filetree.FileTreeViewModel
	*ListContextTrait
}

var (
	_ types.IListContext       = (*WorkingTreeContext)(nil)
	_ types.IFilterableContext = (*WorkingTreeContext)(nil)
)

func NewWorkingTreeContext(c *ContextCommon) *WorkingTreeContext {
	viewModel := filetree.NewFileTreeViewModel(
		func() []*models.File { return c.Model().Files },
		c.Common,
		c.UserConfig().Gui.ShowFileTree,
	)

	var ctx *WorkingTreeContext

	getDisplayStrings := func(_ int, _ int) [][]string {
		// when the panel is collapsed (not the current side panel), show a
		// one-line summary instead of the top of the file list
		if c.Context().CurrentSide().GetWindowName() != ctx.GetWindowName() {
			return [][]string{{style.FgYellow.Sprint(utils.ResolvePlaceholderString(
				c.Tr.DirtyFileCount,
				map[string]string{"count": fmt.Sprint(len(c.Model().Files))},
			))}}
		}

		showFileIcons := icons.IsIconEnabled() && c.UserConfig().Gui.ShowFileIcons
		showNumstat := c.UserConfig().Gui.ShowNumstatInFilesView
		lines := presentation.RenderFileTree(viewModel, c.Model().Submodules, showFileIcons, showNumstat, &c.UserConfig().Gui.CustomIcons, c.UserConfig().Gui.ShowRootItemInFileTree)
		return lo.Map(lines, func(line string, _ int) []string {
			return []string{line}
		})
	}

	ctx = &WorkingTreeContext{
		FileTreeViewModel: viewModel,
		ListContextTrait: &ListContextTrait{
			Context: NewSimpleContext(NewBaseContext(NewBaseContextOpts{
				View:       c.Views().Files,
				WindowName: "files",
				Key:        FILES_CONTEXT_KEY,
				Kind:       types.SIDE_CONTEXT,
				Focusable:  true,
			})),
			ListRenderer: ListRenderer{
				list:              viewModel,
				getDisplayStrings: getDisplayStrings,
			},
			c: c,
		},
	}

	return ctx
}
