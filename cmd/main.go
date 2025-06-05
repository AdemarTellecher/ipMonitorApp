package main

import (
	"fmt"
	"log"
	"time"

	"github.com/AdemarTellecher/ipmonitorapp/internal/controller"
	"github.com/AdemarTellecher/ipmonitorapp/internal/model"
	"github.com/AdemarTellecher/ipmonitorapp/internal/view"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const dbFile = "ipmonitor.db"

func main() {
	repo, err := model.NewRepository(dbFile)
	if err != nil {
		log.Fatalf("Erro ao inicializar o banco: %v", err)
	}
	defer repo.Close()
	ctrl := controller.NewIPController(repo)

	myApp := app.New()

	// Escolha de tema: dark ou light
	isDark := true // Altere para false para tema claro
	if isDark {
		myApp.Settings().SetTheme(theme.DarkTheme())
	} else {
		myApp.Settings().SetTheme(theme.LightTheme())
	}

	var tableData [][]string
	selectedIndex := -1

	refreshTable := func() {}
	updateStatus := func() {}
	addIP := func() {}
	removeIP := func() {}
	onSelect := func(id int) {}

	ui := view.NewMainView(myApp, func() { updateStatus() }, func() { updateStatus() }, func() { addIP() }, func() { removeIP() }, func(id int) { onSelect(id) })

	// Referência à tabela
	ipTable := ui.IPCard.Objects[1].(*widget.Table)

	refreshTable = func() {
		ips, _ := ctrl.ListIPs()
		tableData = make([][]string, len(ips))
		total, on, off := 0, 0, 0
		for i, d := range ips {
			tableData[i] = []string{d.IP, d.Status, "Remover"}
			total++
			if d.Status == "Online" {
				on++
			} else if d.Status == "Offline" {
				off++
			}
		}
		ui.StatusCard.Objects[2].(*widget.Label).SetText("Total de IPs: " + itoa(total))
		ui.StatusCard.Objects[3].(*widget.Label).SetText("Online: " + itoa(on))
		ui.StatusCard.Objects[4].(*widget.Label).SetText("Offline: " + itoa(off))
		ipTable.Length = func() (int, int) { return len(tableData), 3 }
		ipTable.UpdateCell = func(id widget.TableCellID, o fyne.CanvasObject) {
			if id.Row < len(tableData) {
				row := tableData[id.Row]
				if id.Col == 0 {
					label := o.(*fyne.Container).Objects[0].(*widget.Label)
					label.SetText(row[0])
				} else if id.Col == 1 {
					label := o.(*fyne.Container).Objects[1].(*widget.Label)
					label.SetText(row[1])
				} else if id.Col == 2 {
					btn := o.(*fyne.Container).Objects[2].(*widget.Button)
					btn.OnTapped = func() {
						_ = ctrl.RemoveIP(row[0])
						refreshTable()
					}
				}
			}
		}
		ipTable.Refresh()
	}

	updateStatus = func() {
		go func() {
			ctrl.UpdateAllStatuses()
			if drv, ok := fyne.CurrentApp().Driver().(interface{ CallOnMainThread(func()) }); ok {
				drv.CallOnMainThread(func() {
					refreshTable()
				})
			}
		}()
	}

	addIP = func() {
		ip := ui.IPEntry.Text
		if ip == "" {
			return
		}
		err := ctrl.AddIP(ip)
		if err != nil {
			view.ShowError(ui.Window, err.Error())
			return
		}
		ui.IPEntry.SetText("")
		refreshTable()
	}

	removeIP = func() {
		if selectedIndex >= 0 && selectedIndex < len(tableData) {
			ip := tableData[selectedIndex][0]
			_ = ctrl.RemoveIP(ip)
			refreshTable()
			selectedIndex = -1
		}
	}

	onSelect = func(id int) {
		selectedIndex = id
	}

	refreshTable()

	// Atualização automática a cada 1 minuto
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			ctrl.UpdateAllStatuses()
			if drv, ok := fyne.CurrentApp().Driver().(interface{ CallOnMainThread(func()) }); ok {
				drv.CallOnMainThread(func() {
					refreshTable()
				})
			}
		}
	}()

	ui.Window.ShowAndRun()
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
