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
type ViewComponents struct {
	Window          fyne.Window
	TotalValLabel   *widget.Label
	OnlineValLabel  *widget.Label
	OfflineValLabel *widget.Label
	IPEntry         *widget.Entry
	AddBtn          *widget.Button
	UpdateBtn       *widget.Button
	RemoveBtn       *widget.Button
	ImportBtn       *widget.Button
	IPTable         *widget.Table
}

// NewMainView monta a interface com base no novo design moderno refinado
func NewMainView(app fyne.App, refreshList func(), updateStatus func(), addIP func(), removeIP func(), importJSON func(), onSelect func(int)) *ViewComponents {
	w := app.NewWindow("IP Monitor App")

	// ==========================================
	// 1. HEADER MODERNO REFINADO E COMPACTO
	// ==========================================
	titleLabel := widget.NewLabelWithStyle("IP Monitor", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	subTitle := canvas.NewText("v2.0 • Live Ping", color.RGBA{96, 165, 250, 255})
	subTitle.TextSize = 10
	subTitle.TextStyle = fyne.TextStyle{Bold: true}

	brandText := container.NewVBox(titleLabel, subTitle)

	// Ícone em caixinha arredondada estilizada
	iconBg := canvas.NewRectangle(color.RGBA{37, 99, 235, 255})
	iconBg.SetMinSize(fyne.NewSize(32, 32))
	iconBg.CornerRadius = 6
	appIcon := widget.NewIcon(theme.ComputerIcon())
	brandIconBox := container.NewStack(iconBg, container.NewCenter(appIcon))

	brandContainer := container.NewHBox(brandIconBox, brandText)

	// Botão compacto e sutil de alternar tema (LowImportance para não ser azul agressivo)
	isDark := true
	var themeBtn *widget.Button
	themeBtn = widget.NewButtonWithIcon("Tema", theme.ColorPaletteIcon(), func() {
		isDark = !isDark
		if isDark {
			app.Settings().SetTheme(CustomDarkTheme{})
		} else {
			app.Settings().SetTheme(CustomLightTheme{})
		}
		refreshList()
	})
	themeBtn.Importance = widget.LowImportance

	headerBox := container.NewBorder(nil, nil, brandContainer, container.NewCenter(themeBtn))

	headerBg := canvas.NewRectangle(color.RGBA{22, 28, 38, 255})
	headerBg.CornerRadius = 8
	headerCard := container.NewStack(headerBg, container.NewPadded(headerBox))

	// ==========================================
	// 2. CARDS DE MÉTRICAS COMPACTOS COM DESTAQUE
	// ==========================================
	totalValLabel := widget.NewLabelWithStyle("0", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	onlineValLabel := widget.NewLabelWithStyle("0", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	offlineValLabel := widget.NewLabelWithStyle("0", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Total Card (Azul)
	totalTitle := canvas.NewText("TOTAL", color.RGBA{148, 163, 184, 255})
	totalTitle.TextSize = 10
	totalTitle.Alignment = fyne.TextAlignCenter
	totalTitle.TextStyle = fyne.TextStyle{Bold: true}
	totalBox := container.NewVBox(totalTitle, totalValLabel)
	totalBg := canvas.NewRectangle(color.RGBA{26, 34, 48, 255})
	totalBg.CornerRadius = 6
	totalCard := container.NewStack(totalBg, container.NewPadded(totalBox))

	// Online Card (Verde com contraste aprimorado)
	onlineTitle := canvas.NewText("ONLINE", color.RGBA{52, 211, 153, 255})
	onlineTitle.TextSize = 10
	onlineTitle.Alignment = fyne.TextAlignCenter
	onlineTitle.TextStyle = fyne.TextStyle{Bold: true}
	onlineBox := container.NewVBox(onlineTitle, onlineValLabel)
	onlineBg := canvas.NewRectangle(color.RGBA{16, 45, 36, 255})
	onlineBg.CornerRadius = 6
	onlineCard := container.NewStack(onlineBg, container.NewPadded(onlineBox))

	// Offline Card (Vermelho com contraste aprimorado)
	offlineTitle := canvas.NewText("OFFLINE", color.RGBA{248, 113, 113, 255})
	offlineTitle.TextSize = 10
	offlineTitle.Alignment = fyne.TextAlignCenter
	offlineTitle.TextStyle = fyne.TextStyle{Bold: true}
	offlineBox := container.NewVBox(offlineTitle, offlineValLabel)
	offlineBg := canvas.NewRectangle(color.RGBA{48, 24, 30, 255})
	offlineBg.CornerRadius = 6
	offlineCard := container.NewStack(offlineBg, container.NewPadded(offlineBox))

	metricsGrid := container.New(layout.NewGridLayoutWithColumns(3), totalCard, onlineCard, offlineCard)

	// ==========================================
	// 3. BARRA DE ENTRADA UNIFICADA E AÇÕES
	// ==========================================
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("Digite o IP (ex: 192.168.1.1)")
	ipEntry.TextStyle = fyne.TextStyle{Monospace: true}
	ipEntry.OnSubmitted = func(_ string) {
		addIP()
	}

	addBtn := widget.NewButtonWithIcon("Adicionar", theme.ContentAddIcon(), func() {
		addIP()
	})
	addBtn.Importance = widget.HighImportance // Destaque azul primário exclusivo

	inputBar := container.NewBorder(nil, nil, nil, addBtn, ipEntry)

	updateStatusBtn := widget.NewButtonWithIcon("Atualizar Todos", theme.ViewRefreshIcon(), func() {
		updateStatus()
	})
	updateStatusBtn.Importance = widget.MediumImportance

	importBtn := widget.NewButtonWithIcon("Importar JSON", theme.FolderOpenIcon(), func() {
		importJSON()
	})
	importBtn.Importance = widget.LowImportance // Estilo neutro/outline elegante

	removeBtn := widget.NewButtonWithIcon("Remover", theme.DeleteIcon(), func() {
		removeIP()
	})
	removeBtn.Importance = widget.DangerImportance

	toolbar := container.New(layout.NewGridLayoutWithColumns(3), updateStatusBtn, importBtn, removeBtn)

	actionsTitle := canvas.NewText("AÇÕES RÁPIDAS", color.RGBA{148, 163, 184, 255})
	actionsTitle.TextSize = 11
	actionsTitle.TextStyle = fyne.TextStyle{Bold: true}

	actionsContent := container.NewVBox(actionsTitle, inputBar, toolbar)
	actionsBg := canvas.NewRectangle(color.RGBA{22, 28, 38, 255})
	actionsBg.CornerRadius = 8
	actionsCard := container.NewStack(actionsBg, container.NewPadded(actionsContent))

	// ==========================================
	// 4. TABELA MODERNA DE HOSTS MONITORADOS
	// ==========================================
	tableHeaderTitle := canvas.NewText("HOSTS MONITORADOS", color.RGBA{148, 163, 184, 255})
	tableHeaderTitle.TextSize = 11
	tableHeaderTitle.TextStyle = fyne.TextStyle{Bold: true}

	ipHeaderLabel := canvas.NewText("ENDEREÇO IP / HOST", color.RGBA{100, 116, 139, 255})
	ipHeaderLabel.TextSize = 10
	ipHeaderLabel.TextStyle = fyne.TextStyle{Bold: true}

	statusHeaderLabel := canvas.NewText("ESTADO", color.RGBA{100, 116, 139, 255})
	statusHeaderLabel.TextSize = 10
	statusHeaderLabel.TextStyle = fyne.TextStyle{Bold: true}
	statusHeaderLabel.Alignment = fyne.TextAlignCenter

	tableHeaderRow := container.New(
		layout.NewGridLayoutWithColumns(2),
		container.NewPadded(ipHeaderLabel),
		container.NewPadded(statusHeaderLabel),
	)
	tableHeaderBg := canvas.NewRectangle(color.RGBA{16, 20, 28, 255})
	tableHeaderBg.CornerRadius = 4
	tableHeaderBox := container.NewStack(tableHeaderBg, tableHeaderRow)

	ipTable := widget.NewTable(
		func() (int, int) { return 0, 2 },
		func() fyne.CanvasObject {
			// Template de célula da tabela
			bgBadge := canvas.NewRectangle(color.Transparent)
			bgBadge.CornerRadius = 4
			textLabel := widget.NewLabel("")
			textLabel.TextStyle = fyne.TextStyle{Monospace: true}
			return container.NewStack(bgBadge, container.NewCenter(textLabel))
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {},
	)
	ipTable.SetColumnWidth(0, 230)
	ipTable.SetColumnWidth(1, 150)
	ipTable.OnSelected = func(id widget.TableCellID) {
		onSelect(id.Row)
	}

	scrollTable := container.NewScroll(ipTable)
	scrollTable.SetMinSize(fyne.NewSize(400, 290))

	hostsBox := container.NewBorder(
		container.NewVBox(tableHeaderTitle, tableHeaderBox),
		nil, nil, nil,
		scrollTable,
	)
	hostsBg := canvas.NewRectangle(color.RGBA{22, 28, 38, 255})
	hostsBg.CornerRadius = 8
	hostsCard := container.NewStack(hostsBg, container.NewPadded(hostsBox))

	// ==========================================
	// 5. RODAPÉ DE STATUS
	// ==========================================
	footerStatus := canvas.NewText("● Monitoramento automático ativo (1m)", color.RGBA{100, 116, 139, 255})
	footerStatus.TextSize = 10

	footerDb := canvas.NewText("SQLite OK", color.RGBA{52, 211, 153, 255})
	footerDb.TextSize = 10
	footerDb.TextStyle = fyne.TextStyle{Bold: true}

	footerBox := container.NewBorder(nil, nil, footerStatus, footerDb)

	// Layout Completo Integrado
	mainBox := container.NewBorder(
		headerCard,
		footerBox,
		nil,
		nil,
		container.NewVBox(
			metricsGrid,
			actionsCard,
			hostsCard,
		),
	)

	w.SetContent(container.NewPadded(mainBox))
	w.Resize(fyne.NewSize(460, 720))

	return &ViewComponents{
		Window:          w,
		TotalValLabel:   totalValLabel,
		OnlineValLabel:  onlineValLabel,
		OfflineValLabel: offlineValLabel,
		IPEntry:         ipEntry,
		AddBtn:          addBtn,
		UpdateBtn:       updateStatusBtn,
		RemoveBtn:       removeBtn,
		ImportBtn:       importBtn,
		IPTable:         ipTable,
	}
}

// ShowError exibe um diálogo de erro
func ShowError(win fyne.Window, msg string) {
	dialog.NewInformation("Erro", msg, win).Show()
}
