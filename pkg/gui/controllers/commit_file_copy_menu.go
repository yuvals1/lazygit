package controllers

import (
	"path/filepath"

	"github.com/jesseduffield/lazygit/pkg/gui/filetree"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

// commitFileCopyMenuOpts parameterizes openCommitFileCopyMenu over the parts
// that differ between the panels listing commit files (the diff-files panel
// and the files-from-main panel): what the diff items diff against, and where
// "file content" comes from.
type commitFileCopyMenuOpts struct {
	// runs the diff restricted to the given paths and returns its output
	diffForPaths func(paths []string) (string, error)
	// returns the content of the file at the given path
	fileContent func(path string) (string, error)
	// the paths to pass to diffForPaths for the selected node
	pathsForDiff func(node *filetree.CommitFileNode) []string

	singleItemDisabledReason *types.DisabledReason
	itemsDisabledReason      *types.DisabledReason
}

func openCommitFileCopyMenu(c *ControllerCommon, node *filetree.CommitFileNode, opts commitFileCopyMenuOpts) error {
	copyToClipboard := func(content string, toastMessage string) error {
		if err := c.OS().CopyToClipboard(content); err != nil {
			return err
		}
		c.Toast(toastMessage)
		return nil
	}

	copyNameItem := &types.MenuItem{
		Label: c.Tr.CopyFileName,
		OnPress: func() error {
			return copyToClipboard(node.Name(), c.Tr.FileNameCopiedToast)
		},
		DisabledReason: opts.singleItemDisabledReason,
		Keys:           menuKey('n'),
	}
	copyRelativePathItem := &types.MenuItem{
		Label: c.Tr.CopyRelativeFilePath,
		OnPress: func() error {
			return copyToClipboard(node.GetPath(), c.Tr.FilePathCopiedToast)
		},
		DisabledReason: opts.singleItemDisabledReason,
		Keys:           menuKey('p'),
	}
	copyAbsolutePathItem := &types.MenuItem{
		Label: c.Tr.CopyAbsoluteFilePath,
		OnPress: func() error {
			absPath, err := filepath.Abs(node.GetPath())
			if err != nil {
				return err
			}
			return copyToClipboard(absPath, c.Tr.FilePathCopiedToast)
		},
		DisabledReason: opts.singleItemDisabledReason,
		Keys:           menuKey('P'),
	}
	copyFileDiffItem := &types.MenuItem{
		Label: c.Tr.CopySelectedDiff,
		OnPress: func() error {
			diff, err := opts.diffForPaths(opts.pathsForDiff(node))
			if err != nil {
				return err
			}
			return copyToClipboard(diff, c.Tr.FileDiffCopiedToast)
		},
		DisabledReason: opts.singleItemDisabledReason,
		Keys:           menuKey('s'),
	}
	copyAllDiff := &types.MenuItem{
		Label: c.Tr.CopyAllFilesDiff,
		OnPress: func() error {
			diff, err := opts.diffForPaths([]string{"."})
			if err != nil {
				return err
			}
			return copyToClipboard(diff, c.Tr.AllFilesDiffCopiedToast)
		},
		DisabledReason: opts.itemsDisabledReason,
		Keys:           menuKey('a'),
	}
	contentDisabledReason := opts.singleItemDisabledReason
	if contentDisabledReason == nil && node != nil && !node.IsFile() {
		contentDisabledReason = &types.DisabledReason{
			Text:             c.Tr.ErrCannotCopyContentOfDirectory,
			ShowErrorInPanel: true,
		}
	}

	copyFileContentItem := &types.MenuItem{
		Label: c.Tr.CopyFileContent,
		OnPress: func() error {
			content, err := opts.fileContent(node.GetPath())
			if err != nil {
				return err
			}
			return copyToClipboard(content, c.Tr.FileContentCopiedToast)
		},
		DisabledReason: contentDisabledReason,
		Keys:           menuKey('c'),
	}

	return c.Menu(types.CreateMenuOptions{
		Title: c.Tr.CopyToClipboardMenu,
		Items: []*types.MenuItem{
			copyNameItem,
			copyRelativePathItem,
			copyAbsolutePathItem,
			copyFileDiffItem,
			copyAllDiff,
			copyFileContentItem,
		},
	})
}
