package gamedashboard

import (
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/lang"
)

func (m *Model) tavernSleep() {
	req := request.LocationTavernSleep()

	resp, err := req.Send()

	if err != nil {
		m.hideLocationService()
		m.ErrMsg = helper.ConnectionError()
		return
	}

	m.ErrMsg = helper.ExtractError(resp)

	if m.ErrMsg != nil {
		m.hideLocationService()
		return
	}

	m.resetMsg()
	m.SuccessMsg = lang.L("New respawn location set !")
	m.UpdateData()
	m.hideLocationService()
}
