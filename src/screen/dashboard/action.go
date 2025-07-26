package dashboard

import (
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/lang"
	"farental/widget/statusmessage"
)

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
