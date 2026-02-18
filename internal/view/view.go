package view

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ViewComponents agrupa os widgets principais da interface
// para facilitar a manipulação pelo main.go
type ViewComponents struct {
	Window     fyne.Window
	StatusCard *fyne.Container
	IPList     *widget.List
	IPEntry    *widget.Entry
	AddBtn     *widget.Button
	UpdateBtn  *widget.Button
	RemoveBtn  *widget.Button
	ActionCard *fyne.Container
	ActionBox  fyne.CanvasObject
	IPCard     *fyne.Container
	IPCardBox  fyne.CanvasObject
	StatusBox  fyne.CanvasObject
	IPTable    *widget.Table
}

// NewMainView monta e retorna todos os componentes da interface
func NewMainView(app fyne.App, refreshList func(), updateStatus func(), addIP func(), removeIP func(), onSelect func(int)) *ViewComponents {
	w := app.NewWindow("IP Monitor App")

	title := widget.NewLabelWithStyle("IP Monitor - App", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	bgTitle := canvas.NewRectangle(color.RGBA{30, 60, 120, 255})
	bgTitle.SetMinSize(fyne.NewSize(440, 40))
	title.TextStyle = fyne.TextStyle{Bold: true}

	titleBox := container.NewStack(bgTitle, container.NewCenter(title))

	// Criação dos labels de status para evitar duplicidade
	totalLabel := widget.NewLabelWithStyle("Total de IPs: 0", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	onlineLabel := widget.NewLabelWithStyle("Online: 0", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	offlineLabel := widget.NewLabelWithStyle("Offline: 0", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	statusCard := container.NewVBox(
		titleBox,
		widget.NewLabelWithStyle("Resumo do Monitoramento", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		totalLabel,
		onlineLabel,
		offlineLabel,
	)
	statusCardBox := widget.NewCard("", "", container.NewPadded(statusCard))

	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("Digite o IP a ser monitorado")
	addBtn := widget.NewButtonWithIcon("Adicionar IP", theme.ContentAddIcon(), func() { addIP() })
	updateStatusBtn := widget.NewButtonWithIcon("Atualizar Status", theme.ViewRefreshIcon(), func() { updateStatus() })
	removeBtn := widget.NewButtonWithIcon("Remover IP", theme.DeleteIcon(), func() { removeIP() })

	btnBox := container.New(layout.NewGridLayoutWithColumns(3), addBtn, updateStatusBtn, removeBtn)
	actionCard := container.NewVBox(
		widget.NewLabelWithStyle("Ações Rápidas", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		ipEntry,
		btnBox,
	)
	actionCardBox := widget.NewCard("", "", container.NewPadded(actionCard))

	// --- NOVA LISTA DE IPs COMO TABELA ---
	ipTable := widget.NewTable(
		func() (int, int) { return 0, 2 }, // linhas, colunas (IP, Status)
		func() fyne.CanvasObject {
			// Usa GridLayout com 2 colunas para distribuir espaço corretamente
			return container.New(
				layout.NewGridLayoutWithColumns(2),
				widget.NewLabel(""), // IP
				widget.NewLabel(""), // Status
			)
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {},
	)
	ipTable.SetColumnWidth(0, 220) // IP
	ipTable.SetColumnWidth(1, 120) // Status
	ipTable.OnSelected = func(id widget.TableCellID) {
		onSelect(id.Row)
	}

	// Envolve a tabela em um scroll com altura definida
	scrollTable := container.NewScroll(ipTable)
	scrollTable.SetMinSize(fyne.NewSize(440, 300))

	ipCard := container.NewVBox(
		widget.NewLabelWithStyle("IPs Monitorados", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("IPs STATUS", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		scrollTable,
	)
	ipCardBox := widget.NewCard("", "", container.NewPadded(ipCard))

	// Botão para alternar tema
	isDark := false
	themeBtn := widget.NewButtonWithIcon("Alternar Tema", theme.ColorPaletteIcon(), func() {
		isDark = !isDark
		if isDark {
			app.Settings().SetTheme(CustomDarkTheme{})
		} else {
			app.Settings().SetTheme(CustomLightTheme{})
		}
		refreshList()
	})

	// Adiciona o botão de tema ao topo
	mainBox := container.NewVBox(
		container.NewHBox(layout.NewSpacer(), themeBtn),
		statusCardBox,
		actionCardBox,
		ipCardBox,
	)
	// Remove o background customizado para permitir que o tema do Fyne seja aplicado corretamente
	w.SetContent(mainBox)
	w.Resize(fyne.NewSize(460, 700))

	return &ViewComponents{
		Window:     w,
		StatusCard: statusCard,
		IPList:     nil, // widget.List removido
		IPEntry:    ipEntry,
		AddBtn:     addBtn,
		UpdateBtn:  updateStatusBtn,
		RemoveBtn:  removeBtn,
		ActionCard: actionCard,
		ActionBox:  actionCardBox,
		IPCard:     ipCard,
		IPCardBox:  ipCardBox,
		StatusBox:  statusCardBox,
		IPTable:    ipTable,
	}
}

// ShowError exibe um diálogo de erro
func ShowError(win fyne.Window, msg string) {
	dialog.NewInformation("Erro", msg, win).Show()
}

// go install fyne.io/fyne/v2/cmd/fyne@latest
