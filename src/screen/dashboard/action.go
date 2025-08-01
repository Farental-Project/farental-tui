package dashboard

import (
	"errors"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/lang"
	"farental/widget/statusmessage"
	"log"
)

func (s *Screen) runningTaskError() {
	if context.RunningTask.IsRunning {
		s.statusMessage.SetError(
			errors.New(lang.L("A task is currently running.")))
	} else {
		s.statusMessage.SetError(
			errors.New(lang.L("Please claim your reward first.")))
	}
}

func (s *Screen) tavernSleep() {
	_, err := helper.SendRequest(request.LocationTavernSleep())

	if err != nil {
		// TODO : Hide location service
		s.statusMessage.SetError(err)
		return
	}

	s.statusMessage.SetMessage(
		lang.L("New respawn location set !"),
		statusmessage.SuccessMessage,
	)

	// TODO : Update data and hide location service
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
