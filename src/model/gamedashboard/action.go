package gamedashboard

import (
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/lang"
)

func (m *Model) tavernSleep() {
	_, err := helper.SendRequest(request.LocationTavernSleep())

	if err != nil {
		m.hideLocationService()
		m.ErrMsg = err
		return
	}

	m.resetMsg()
	m.SuccessMsg = lang.L("New respawn location set !")
	m.UpdateData()
	m.hideLocationService()
}
