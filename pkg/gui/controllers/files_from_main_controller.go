package controllers

import (
	"os"

	"github.com/jesseduffield/lazygit/pkg/gui/context"
	"github.com/jesseduffield/lazygit/pkg/gui/filetree"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/samber/lo"
)

// FilesFromMainController drives the tab listing the files that differ between
// the working tree and the merge base with the main branch.
type FilesFromMainController struct {
	baseController
	*ListControllerTrait[*filetree.CommitFileNode]
	c *ControllerCommon
}

var _ types.IController = &FilesFromMainController{}

func NewFilesFromMainController(
	c *ControllerCommon,
) *FilesFromMainController {
	return &FilesFromMainController{
		baseController: baseController{},
		ListControllerTrait: NewListControllerTrait(
			c,
			c.Contexts().FilesFromMain,
			c.Contexts().FilesFromMain.GetSelected,
			c.Contexts().FilesFromMain.GetSelectedItems,
		),
		c: c,
	}
}

func (self *FilesFromMainController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return []*types.Binding{
		{
			Keys:              opts.GetKeys(opts.Config.Universal.Select),
			Handler:           self.withItem(self.enter),
			GetDisabledReason: self.require(self.singleItemSelected()),
			Description:       self.c.Tr.Enter,
		},
		{
			Keys:              opts.GetKeys(opts.Config.Universal.GoInto),
			Handler:           self.withItem(self.enter),
			GetDisabledReason: self.require(self.singleItemSelected()),
		},
		{
			Keys:              opts.GetKeys(opts.Config.Universal.Edit),
			Handler:           self.withItems(self.edit),
			GetDisabledReason: self.require(self.itemsSelected(self.canEditFiles)),
			Description:       self.c.Tr.Edit,
		},
		{
			Keys:              opts.GetKeys(opts.Config.Universal.OpenFile),
			Handler:           self.withItem(self.open),
			GetDisabledReason: self.require(self.singleItemSelected()),
			Description:       self.c.Tr.OpenFile,
		},
		{
			Keys:        opts.GetKeys(opts.Config.Files.ToggleTreeView),
			Handler:     self.toggleTreeView,
			Description: self.c.Tr.ToggleTreeView,
		},
		{
			Keys:        opts.GetKeys(opts.Config.Files.CopyFileInfoToClipboard),
			Handler:     self.openCopyMenu,
			Description: self.c.Tr.CopyToClipboardMenu,
			OpensMenu:   true,
		},
	}
}

func (self *FilesFromMainController) GetOnRenderToMain() func() {
	return func() {
		node := self.context().GetSelected()
		baseRef := self.context().GetBaseRef()
		if node == nil || baseRef == "" {
			self.c.RenderToMainViews(types.RefreshMainOpts{
				Pair: self.c.MainViewPairs().Normal,
				Main: &types.ViewUpdateOpts{
					Title: self.c.Tr.DiffTitle,
					Task:  types.NewRenderStringTask(self.c.Tr.NoChangedFiles),
				},
			})
			return
		}

		cmdObj := self.c.Git().WorkingTree.ShowWorktreeDiffAgainstRefCmdObj(baseRef, []string{node.GetPath()}, false)
		task := types.NewRunPtyTask(cmdObj.GetCmd())

		self.c.RenderToMainViews(types.RefreshMainOpts{
			Pair: self.c.MainViewPairs().Normal,
			Main: &types.ViewUpdateOpts{
				Title:    self.c.Tr.DiffTitle,
				SubTitle: self.c.Helpers().Diff.IgnoringWhitespaceSubTitle(),
				Task:     task,
			},
		})
	}
}

func (self *FilesFromMainController) enter(node *filetree.CommitFileNode) error {
	if node.File == nil {
		self.context().CommitFileTreeViewModel.ToggleCollapsed(node.GetInternalPath())
		self.c.PostRefreshUpdate(self.context())
		return nil
	}

	return self.c.Helpers().Files.EditFiles([]string{node.GetPath()})
}

func (self *FilesFromMainController) openCopyMenu() error {
	return openCommitFileCopyMenu(self.c, self.context().GetSelected(), commitFileCopyMenuOpts{
		diffForPaths: func(paths []string) (string, error) {
			return self.c.Git().WorkingTree.
				ShowWorktreeDiffAgainstRefCmdObj(self.context().GetBaseRef(), paths, true).
				RunWithOutput()
		},
		fileContent: func(path string) (string, error) {
			content, err := os.ReadFile(path)
			return string(content), err
		},
		pathsForDiff: func(node *filetree.CommitFileNode) []string {
			return diffPathsForNode(
				node.Raw(), self.context().GetRoot().Raw(), self.c.Model().FilesFromMain, self.context().IsFiltering())
		},
		singleItemDisabledReason: self.require(self.singleItemSelected())(),
		itemsDisabledReason:      self.require(self.itemsSelected())(),
	})
}

func (self *FilesFromMainController) open(node *filetree.CommitFileNode) error {
	return self.c.Helpers().Files.OpenFile(node.GetPath())
}

func (self *FilesFromMainController) edit(nodes []*filetree.CommitFileNode) error {
	return self.c.Helpers().Files.EditFiles(lo.FilterMap(nodes,
		func(node *filetree.CommitFileNode, _ int) (string, bool) {
			return node.GetPath(), node.IsFile()
		}))
}

func (self *FilesFromMainController) canEditFiles(nodes []*filetree.CommitFileNode) *types.DisabledReason {
	if lo.NoneBy(nodes, func(node *filetree.CommitFileNode) bool { return node.IsFile() }) {
		return &types.DisabledReason{
			Text:             self.c.Tr.ErrCannotEditDirectory,
			ShowErrorInPanel: true,
		}
	}

	return nil
}

func (self *FilesFromMainController) toggleTreeView() error {
	self.context().CommitFileTreeViewModel.ToggleShowTree()
	self.c.PostRefreshUpdate(self.context())
	return nil
}

// the panel renders a one-line summary while collapsed (unfocused) and the full
// file tree while focused, so re-render whenever focus changes. These live on
// the controller because context-constructor hooks are cleared when controllers
// are attached (ClearAllAttachedControllerFunctions).
func (self *FilesFromMainController) GetOnFocus() func(types.OnFocusOpts) {
	return func(types.OnFocusOpts) {
		self.context().HandleRender()
	}
}

func (self *FilesFromMainController) GetOnFocusLost() func(types.OnFocusLostOpts) {
	return func(types.OnFocusLostOpts) {
		self.context().HandleRender()
	}
}

func (self *FilesFromMainController) context() *context.FilesFromMainContext {
	return self.c.Contexts().FilesFromMain
}
