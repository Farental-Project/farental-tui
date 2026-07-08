package main

import (
	"embed"
	"farental/core/data"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/config"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/style"
	ftheme "farental/internal/theme"
	"farental/screen"
	"farental/screen/accountcreation"
	"farental/screen/activity"
	"farental/screen/bank"
	"farental/screen/charactercreation"
	"farental/screen/characterselection"
	"farental/screen/charactersheet"
	"farental/screen/chat"
	"farental/screen/craft"
	"farental/screen/dashboard"
	"farental/screen/fight"
	"farental/screen/fighthistory"
	"farental/screen/inventory"
	"farental/screen/locationinfo"
	"farental/screen/login"
	"farental/screen/mailbox"
	"farental/screen/maileditor"
	"farental/screen/mailreader"
	"farental/screen/npc"
	"farental/screen/scripteditor"
	"farental/screen/scriptexplorer"
	"farental/screen/sendfeedback"
	"farental/screen/shop"
	"farental/screen/travel"
	"farental/screen/usersettings"
	"fmt"
	"log"
	"strings"

	"github.com/halsten-dev/orvyn"

	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/spf13/viper"

	tea "github.com/charmbracelet/bubbletea"
)

//go:embed translations
var translations embed.FS

func main() {
	var err error

	// Debug
	f, err := tea.LogToFile("debug_farental.log", "debug")

	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	config.Init()
	context.Init()
	request.Init(context.Client)
	lokyn.Init()
	err = lokyn.AddTranslationFS(translations, "translations")

	if err != nil {
		log.Fatal(err)
	}

	lokyn.SetLanguage(viper.GetString("language"))

	// Check version
	reqVer := request.VersionGet()
	version, err := helper.Fetch[api.DbVersion](reqVer)

	if err != nil {
		fmt.Println(lokyn.L("Cannot verify server version. Please retry later."))
		return
	}

	if !strings.HasPrefix(config.VERSION, version.ClientTui) {
		fmt.Println(lokyn.L("Your client version is not aligned with the server. Please update it."))
		fmt.Println(lokyn.L("Visit https://www.farental.ch for more informations."))
		return
	}

	// Other init
	keybind.Init()

	bubblehelp.Init()

	data.InitTargets()

	// Orvyn
	orvyn.Init()

	orvyn.SetTheme(ftheme.GetTheme(viper.GetString("theme")))

	style.InitHelpStyle()

	keybind.InitContexts()

	orvyn.RegisterScreen(screen.IDLogin, login.New())
	orvyn.RegisterScreen(screen.IDUserSettings, usersettings.New())
	orvyn.RegisterScreen(screen.IDAccountCreation, accountcreation.New())
	orvyn.RegisterScreen(screen.IDCharacterSelection, characterselection.New())
	orvyn.RegisterScreen(screen.IDCharacterCreation, charactercreation.New())
	orvyn.RegisterScreen(screen.IDDashBoard, dashboard.New())
	orvyn.RegisterScreen(screen.IDTravel, travel.New())
	orvyn.RegisterScreen(screen.IDActivity, activity.New())
	orvyn.RegisterScreen(screen.IDFight, fight.New())
	orvyn.RegisterScreen(screen.IDFightHistory, fighthistory.New())
	orvyn.RegisterScreen(screen.IDCraft, craft.New())
	orvyn.RegisterScreen(screen.IDChat, chat.New())
	orvyn.RegisterScreen(screen.IDInventory, inventory.New())
	orvyn.RegisterScreen(screen.IDCharacterSheet, charactersheet.New())
	orvyn.RegisterScreen(screen.IDMailBox, mailbox.New())
	orvyn.RegisterScreen(screen.IDMailReader, mailreader.New())
	orvyn.RegisterScreen(screen.IDMailEditor, maileditor.New())
	orvyn.RegisterScreen(screen.IDScriptExplorer, scriptexplorer.New())
	orvyn.RegisterScreen(screen.IDScriptEditor, scripteditor.New())
	orvyn.RegisterScreen(screen.IDBank, bank.New())
	orvyn.RegisterScreen(screen.IDNpc, npc.New())
	orvyn.RegisterScreen(screen.IDShop, shop.New())
	orvyn.RegisterScreen(screen.IDLocationInfo, locationinfo.New())
	orvyn.RegisterScreen(screen.IDSendFeedback, sendfeedback.New())

	p := tea.NewProgram(&App{}, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
