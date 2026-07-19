package runningtask

import (
	"farental/art"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
)

type Style struct {
	Widget        lipgloss.Style
	NoTask        lipgloss.Style
	TaskRunning   lipgloss.Style
	SpinnerWidget lipgloss.Style
	Spinner       lipgloss.Style
}

type Widget struct {
	orvyn.BaseWidget

	Style Style

	spinner spinner.Model

	task *api.TaskResponse
}

func New() *Widget {
	w := new(Widget)

	t := orvyn.GetTheme()

	w.BaseWidget = orvyn.NewBaseWidget()

	w.task = nil

	w.Style = Style{
		Widget: t.Style(theme.BlurredWidgetStyleID).
			AlignHorizontal(lipgloss.Center),
		NoTask:        t.Style(theme.DimTextStyleID),
		TaskRunning:   t.Style(theme.TitleStyleID),
		SpinnerWidget: t.Style(theme.BlurredWidgetStyleID),
		Spinner:       t.Style(theme.TitleStyleID),
	}

	w.spinner = spinner.New()
	w.spinner.Spinner = spinner.Spinner{
		Frames: art.WaitSpinner,
		FPS:    time.Second / 9,
	}
	w.spinner.Style = w.Style.Spinner

	return w
}

func (w *Widget) Init() tea.Cmd {
	return w.spinner.Tick
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		w.spinner, cmd = w.spinner.Update(msg)

		return cmd
	}

	return nil
}

func (w *Widget) Render() string {
	var b strings.Builder

	contentSize := w.GetContentSize()

	w.Style.Widget = w.Style.Widget.Width(contentSize.Width)

	if w.task != nil {
		b.WriteString(w.Style.TaskRunning.Render(
			w.task.Title,
		))
		b.WriteString("\n")

		if w.task.RemainingTimeHours > 0 {
			fmt.Fprintf(&b, "%s : %s", lokyn.L("Remaining time"),
				helper.HoursDecFormat(w.task.RemainingTimeHours))
			b.WriteString("\n")
			b.WriteString(w.Style.SpinnerWidget.Render(w.spinner.View()))
		} else {
			b.WriteString(lokyn.L("Completed! Waiting for claim!"))
		}
	} else {
		b.WriteString(w.Style.NoTask.Render(
			lokyn.L("No running task")),
		)
	}

	return w.Style.Widget.Render(b.String())
}

func (w *Widget) UpdateData(task *api.TaskResponse) {
	w.task = task
}

func (w *Widget) GetData() *api.TaskResponse {
	return w.task
}

func (w *Widget) GetMinSize() orvyn.Size {
	if w.task != nil {
		return orvyn.NewSize(10, 7)
	} else {
		return orvyn.NewSize(10, 3)
	}
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return w.GetMinSize()
}

func (w *Widget) RefreshCurrentCharacter() {
	task, err := context.RefreshRunningTask(w.GetData())

	if err != nil {
		log.Println(err)
	}

	w.UpdateData(task)
}

func (w *Widget) RefreshInspectCharacter(characterID uint) {
	task, err := helper.Fetch[api.TaskResponse](request.TaskInspect(characterID))

	if err != nil {
		log.Println(err)
	}

	w.UpdateData(task)
}
