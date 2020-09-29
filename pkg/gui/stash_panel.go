package gui

import (
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/models"
)

// list panel functions

func (gui *Gui) getSelectedStashEntry() *models.StashEntry {
	selectedLine := gui.State.Panels.Stash.SelectedLineIdx
	if selectedLine == -1 {
		return nil
	}

	return gui.State.StashEntries[selectedLine]
}

func (gui *Gui) handleStashEntrySelect() error {
	var task updateTask
	stashEntry := gui.getSelectedStashEntry()
	if stashEntry == nil {
		task = gui.createRenderStringTask(gui.Tr.SLocalize("NoStashEntries"))
	} else {
		cmd := gui.OSCommand.ExecutableFromString(
			gui.GitCommand.ShowStashEntryCmdStr(stashEntry.Index),
		)
		task = gui.createRunPtyTask(cmd)
	}

	return gui.refreshMainViews(refreshMainOpts{
		main: &viewUpdateOpts{
			title: "Stash",
			task:  task,
		},
	})
}

func (gui *Gui) refreshStashEntries() error {
	gui.State.StashEntries = gui.GitCommand.GetStashEntries(gui.State.Modes.Filtering.Path)

	return gui.Contexts.Stash.Context.HandleRender()
}

// specific functions

func (gui *Gui) handleStashApply(g *gocui.Gui, v *gocui.View) error {
	skipStashWarning := gui.Config.GetUserConfig().GetBool("gui.skipStashWarning")

	apply := func() error {
		return gui.stashDo("apply")
	}

	if skipStashWarning {
		return apply()
	}

	return gui.ask(askOpts{
		title:  gui.Tr.SLocalize("StashApply"),
		prompt: gui.Tr.SLocalize("SureApplyStashEntry"),
		handleConfirm: func() error {
			return apply()
		},
	})
}

func (gui *Gui) handleStashPop(g *gocui.Gui, v *gocui.View) error {
	skipStashWarning := gui.Config.GetUserConfig().GetBool("gui.skipStashWarning")

	pop := func() error {
		return gui.stashDo("pop")
	}

	if skipStashWarning {
		return pop()
	}

	return gui.ask(askOpts{
		title:  gui.Tr.SLocalize("StashPop"),
		prompt: gui.Tr.SLocalize("SurePopStashEntry"),
		handleConfirm: func() error {
			return pop()
		},
	})
}

func (gui *Gui) handleStashDrop(g *gocui.Gui, v *gocui.View) error {
	return gui.ask(askOpts{
		title:  gui.Tr.SLocalize("StashDrop"),
		prompt: gui.Tr.SLocalize("SureDropStashEntry"),
		handleConfirm: func() error {
			return gui.stashDo("drop")
		},
	})
}

func (gui *Gui) stashDo(method string) error {
	stashEntry := gui.getSelectedStashEntry()
	if stashEntry == nil {
		errorMessage := gui.Tr.TemplateLocalize(
			"NoStashTo",
			Teml{
				"method": method,
			},
		)
		return gui.createErrorPanel(errorMessage)
	}
	if err := gui.GitCommand.StashDo(stashEntry.Index, method); err != nil {
		return gui.surfaceError(err)
	}
	return gui.refreshSidePanels(refreshOptions{scope: []int{STASH, FILES}})
}

func (gui *Gui) handleStashSave(stashFunc func(message string) error) error {
	if len(gui.trackedFiles()) == 0 && len(gui.stagedFiles()) == 0 {
		return gui.createErrorPanel(gui.Tr.SLocalize("NoTrackedStagedFilesStash"))
	}
	return gui.prompt(gui.Tr.SLocalize("StashChanges"), "", func(stashComment string) error {
		if err := stashFunc(stashComment); err != nil {
			return gui.surfaceError(err)
		}
		return gui.refreshSidePanels(refreshOptions{scope: []int{STASH, FILES}})
	})
}

func (gui *Gui) handleViewStashFiles() error {
	stashEntry := gui.getSelectedStashEntry()
	if stashEntry == nil {
		return nil
	}

	return gui.switchToCommitFilesContext(stashEntry.RefName(), false, gui.Contexts.Stash.Context, "stash")
}
