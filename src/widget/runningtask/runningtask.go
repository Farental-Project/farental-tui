package runningtask

import (
	"farental/art"
	"farental/internal/context"
	"farental/internal/helper"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"strings"
	"time"
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
}

func New() *Widget {
	w := new(Widget)

	t := orvyn.GetTheme()

	w.BaseWidget = orvyn.NewBaseWidget()

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

	size := w.GetSize()

	w.Style.Widget = w.Style.Widget.Width(size.Width - 2)

	if context.RunningTask != nil {
		b.WriteString(w.Style.TaskRunning.Render(
			context.RunningTask.Title,
		))
		b.WriteString("\n")

		if context.RunningTask.RemainingTimeHours > 0 {
			b.WriteString(fmt.Sprintf("%s : %s", lokyn.L("Remaining time"),
				helper.HoursDecFormat(context.RunningTask.RemainingTimeHours)))
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

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.GetRenderSize(w.Style.Widget, " ")
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return w.GetMinSize()
}
