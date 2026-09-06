package presentation

import (
	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/gui/presentation/icons"
	"github.com/jesseduffield/lazygit/pkg/gui/style"
	"github.com/jesseduffield/lazygit/pkg/i18n"
	"github.com/jesseduffield/lazygit/pkg/theme"
	"github.com/jesseduffield/lazygit/pkg/utils"
	"github.com/samber/lo"
)

func GetWorktreeDisplayStrings(tr *i18n.TranslationSet, worktrees []*models.Worktree) [][]string {
	return lo.Map(worktrees, func(worktree *models.Worktree, _ int) []string {
		return GetWorktreeDisplayString(
			tr,
			worktree)
	})
}

func GetWorktreeDisplayString(tr *i18n.TranslationSet, worktree *models.Worktree) []string {
	textStyle := theme.DefaultTextColor

	recency := ""
	recencyColor := style.FgCyan
	if worktree.IsCurrent {
		recency = "  *"
		recencyColor = style.FgGreen
	} else if worktree.LastActivityUnix > 0 {
		recency = utils.UnixToTimeAgo(worktree.LastActivityUnix)
	}

	icon := icons.IconForWorktree(false)
	if worktree.IsPathMissing {
		textStyle = style.FgRed
		icon = icons.IconForWorktree(true)
	}

	res := []string{}
	res = append(res, recencyColor.Sprint(recency))
	if icons.IsIconEnabled() {
		res = append(res, textStyle.Sprint(icon))
	}

	name := worktree.Name
	if worktree.IsMain {
		name += " " + tr.MainWorktree
	}
	if worktree.IsPathMissing && !icons.IsIconEnabled() {
		name += " " + tr.MissingWorktree
	}
	res = append(res, textStyle.Sprint(name))
	return res
}
