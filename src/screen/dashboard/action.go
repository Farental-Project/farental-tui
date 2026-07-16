package dashboard

import (
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/screen/dialog/popup"

	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
)

func (s *Screen) tavernSleep() {
	_, err := helper.SendRequest(request.LocationTavernSleep())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.statusMessage.SetMessage(
		lokyn.L("New respawn location set !"),
		statusmessage.SuccessMessage,
	)

	s.updateData()
}

func (s *Screen) tavernRegen() {
	_, err := helper.SendRequest(request.LocationTavernRegen())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.statusMessage.SetMessage(
		lokyn.L("HP and MP fully regenerated !"),
		statusmessage.SuccessMessage,
	)

	s.updateData()
}

func (s *Screen) claim() {
	if context.RunningTask == nil {
		return
	}

	if context.RunningTask.IsRunning {
		orvyn.OpenDialog("earlyClaimConfirm", popup.NewYesNo(
			lokyn.L("Are you sure you want to claim the current unfinished task? Rewards might be lost."),
		), nil)

		return
	}

	s.doClaim()
}

func (s *Screen) doClaim() {
	_, err := helper.SendRequest(request.TaskClaim())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	context.RunningTask = nil
	s.updateData()
}
