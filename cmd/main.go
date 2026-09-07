package main

import (
	"fmt"
	"log"
	"time"

	"image/color"

	"github.com/AdemarTellecher/ipmonitorapp/internal/controller"
	"github.com/AdemarTellecher/ipmonitorapp/internal/model"
	"github.com/AdemarTellecher/ipmonitorapp/internal/view"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
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
	myApp.Settings().SetTheme(view.CustomDarkTheme{})

	var tableData [][]string
	selectedIndex := -1

	refreshTable := func() {}
	updateStatus := func() {}
	addIP := func() {}
	removeIP := func() {}
	importJSON := func() {}
	onSelect := func(id int) {}

	ui := view.NewMainView(myApp, func() { updateStatus() }, func() { updateStatus() }, func() { addIP() }, func() { removeIP() }, func() { importJSON() }, func(id int) { onSelect(id) })

	// Referência à tabela
	ipTable := ui.IPTable

	// Desabilita o botão Remover inicialmente
	ui.RemoveBtn.Disable()

	refreshTable = func() {
		ips, _ := ctrl.ListIPs()
		tableData = make([][]string, len(ips))
		total, on, off := 0, 0, 0
		for i, d := range ips {
			tableData[i] = []string{d.IP, d.Status}
			total++
			if d.Status == "Online" {
				on++
			} else if d.Status == "Offline" {
				off++
			}
		}

		// Atualiza labels de estatísticas nos novos Cards
		ui.TotalValLabel.SetText(itoa(total))
		ui.OnlineValLabel.SetText(itoa(on))
		ui.OfflineValLabel.SetText(itoa(off))

		ipTable.Length = func() (int, int) { return len(tableData), 2 }
		ipTable.UpdateCell = func(id widget.TableCellID, o fyne.CanvasObject) {
			if id.Row < len(tableData) {
				row := tableData[id.Row]
				cellContainer := o.(*fyne.Container)
				bg := cellContainer.Objects[0].(*canvas.Rectangle)
				padded := cellContainer.Objects[1].(*fyne.Container)
				label := padded.Objects[0].(*widget.Label)

				if id.Col == 0 {
					label.SetText(row[0])
					label.Alignment = fyne.TextAlignLeading
					bg.FillColor = color.Transparent
				} else if id.Col == 1 {
					status := row[1]
					if status == "Online" {
						label.SetText("● Online")
						label.Alignment = fyne.TextAlignCenter
						bg.FillColor = color.RGBA{16, 185, 129, 35} // badge verde sutil
					} else if status == "Offline" {
						label.SetText("✕ Offline")
						label.Alignment = fyne.TextAlignCenter
						bg.FillColor = color.RGBA{239, 68, 68, 35} // badge vermelho sutil
					} else {
						label.SetText("? " + status)
						label.Alignment = fyne.TextAlignCenter
						bg.FillColor = color.Transparent
					}
				}
				cellContainer.Refresh()
			}
		}
		ipTable.Refresh()
		// Reseta seleção e desabilita botão ao atualizar tabela
		selectedIndex = -1
		ui.RemoveBtn.Disable()
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

	importJSON = func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, ui.Window)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			count, err := ctrl.ImportFromJSON(reader)
			if err != nil {
				dialog.ShowError(err, ui.Window)
				return
			}

			refreshTable()
			msg := fmt.Sprintf("%d novo(s) IP(s) importado(s) com sucesso!", count)
			if count == 0 {
				msg = "Nenhum novo IP para importar (todos os IPs do arquivo já existiam no banco)."
			}
			dialog.ShowInformation("Importação Concluída", msg, ui.Window)
		}, ui.Window)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
		fd.Show()
	}

	onSelect = func(id int) {
		selectedIndex = id
		// Habilita o botão quando um item é selecionado
		ui.RemoveBtn.Enable()
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
