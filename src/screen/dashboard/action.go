package dashboard

import (
	"errors"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/widget/statusmessage"
	"github.com/halsten-dev/lokyn"
	"log"
)

func (s *Screen) runningTaskError() {
	if context.RunningTask.IsRunning {
		s.statusMessage.SetError(
			errors.New(lokyn.L("A task is currently running.")))
	} else {
		s.statusMessage.SetError(
			errors.New(lokyn.L("Please claim your reward first.")))
	}
}

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

func (s *Screen) claim() {
	if context.RunningTask == nil {
		return
	}

	if context.RunningTask.IsRunning {
		s.runningTaskError()
	}

	_, err := helper.SendRequest(request.TaskClaim())

	if err != nil {
		log.Println(err)
		return
	}

	context.RunningTask = nil
	s.updateData()
}
