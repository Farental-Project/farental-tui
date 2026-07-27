package context

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"fmt"

	"github.com/halsten-dev/lokyn"
)

// appName prefixes the terminal title so the window is identifiable among
// other terminals.
const appName = "Farental"

// terminalTitle holds the title currently written to the terminal, so it is
// only rewritten when the task state actually changes. It starts empty so the
// first refresh always writes, taking the title over from the shell.
var terminalTitle string

// RefreshCharacterInfo fetches the current currency amount and, when fresh is
// true or nothing is cached yet, character info from the server. It updates
// CharacterID and CharacterInfo, and — on a fresh fetch that crossed into a
// new location — clears ChatContent, mirroring the dashboard's pre-existing
// behavior. The caller feeds the returned values into a characterinfo.Widget
// via its own UpdateData call.
func RefreshCharacterInfo(fresh bool) (*api.CharacterInfoResponse, int, error) {
	info := CharacterInfo

	if fresh || info == nil {
		fetched, err := helper.Fetch[api.CharacterInfoResponse](request.CharacterGetInfo())

		if err != nil {
			return nil, 0, err
		}

		if info == nil || info.Location.ID != fetched.Location.ID {
			ChatContent = make([]string, 0)
		}

		CharacterID = fetched.ID
		CharacterInfo = fetched
		info = fetched
	}

	currencyResp, err := helper.Fetch[api.CurrencyResponse](
		request.CharacterGetCurrencyAmount(api.Grynars))

	if err != nil {
		return nil, 0, err
	}

	return info, currencyResp.Amount, nil
}

// RefreshHaveUnreadMail fetches if the current character have unread mail.
func RefreshHaveUnreadMail() bool {
	_, err := helper.SendRequest(request.MailHaveUnread())

	if err != nil {
		return false
	}

	return true
}

// RefreshRunningTask fetches the player's current running task, if any, and
// updates RunningTask. No widget update call is needed afterwards —
// runningtask.Widget reads RunningTask directly in its Render(). Keeps the
// terminal title in sync with the task state and rings the terminal bell
// exactly once, the moment the task transitions from running to claimable,
// regardless of which screen's ticker triggered this refresh.
func RefreshRunningTask(task *api.TaskResponse) (*api.TaskResponse, error) {
	wasRunning := task != nil && task.RemainingTimeHours > 0

	resp, err := helper.SendRequest(request.TaskGetRunning())

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() == 404 {
		refreshTerminalTitle(nil)
		return nil, nil
	}

	runningTask := resp.Result().(*api.TaskResponse)

	refreshTerminalTitle(runningTask)

	if wasRunning && runningTask.RemainingTimeHours <= 0 {
		helper.Notify(appName, taskTitleSuffix(runningTask))
		helper.Bell()
	}

	return runningTask, nil
}

// ResetTerminalTitle puts the plain application name in the terminal title,
// taking it over from whatever the shell left there.
func ResetTerminalTitle() {
	refreshTerminalTitle(nil)
}

// taskTitleSuffix describes the given task the way it reads after the
// application name in the terminal title, and doubles as the body of the
// completion notification. A nil task means no task is running, and has no
// description.
func taskTitleSuffix(task *api.TaskResponse) string {
	if task == nil {
		return ""
	}

	if task.RemainingTimeHours > 0 {
		return task.Title
	}

	return fmt.Sprintf("[%s] %s", lokyn.L("Ready"), task.Title)
}

// refreshTerminalTitle puts the state of the given task in the terminal title,
// so the player can tell from the window title or the taskbar entry alone what
// the character is doing and whether a reward is waiting. A nil task means no
// task is running. The title is only written when it actually changes.
func refreshTerminalTitle(task *api.TaskResponse) {
	title := appName

	if suffix := taskTitleSuffix(task); suffix != "" {
		title = fmt.Sprintf("%s - %s", appName, suffix)
	}

	if title == terminalTitle {
		return
	}

	terminalTitle = title

	helper.SetTerminalTitle(title)
}
