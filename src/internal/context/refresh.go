package context

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"fmt"
)

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

// RefreshRunningTask fetches the player's current running task, if any, and
// updates RunningTask. No widget update call is needed afterwards —
// runningtask.Widget reads RunningTask directly in its Render(). Rings the
// terminal bell exactly once, the moment the task transitions from running
// to claimable, regardless of which screen's ticker triggered this refresh.
func RefreshRunningTask() error {
	wasRunning := RunningTask != nil && RunningTask.RemainingTimeHours > 0

	resp, err := helper.SendRequest(request.TaskGetRunning())

	if err != nil {
		return err
	}

	if resp.StatusCode() == 404 {
		RunningTask = nil
		return nil
	}

	RunningTask = resp.Result().(*api.TaskResponse)

	if wasRunning && RunningTask.RemainingTimeHours <= 0 {
		fmt.Print("\a")
	}

	return nil
}
