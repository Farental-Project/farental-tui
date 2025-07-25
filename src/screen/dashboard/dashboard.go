package dashboard

import (
	"farental/internal/orvyn"
	"farental/internal/orvyn/layout"
	"farental/widget/errormessage"
	"farental/widget/help"
	"farental/widget/runningtask"
	"farental/widget/simplelogviewer"
	"time"
)

type Screen struct {
	orvyn.BaseScreen

	runningTask *runningtask.Widget

	logEvent *simplelogviewer.Widget

	logChat *simplelogviewer.Widget

	logCharacters *simplelogviewer.Widget

	help *help.Widget

	error *errormessage.Widget

	lastEventLogTimestamp time.Time

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	return s
}
