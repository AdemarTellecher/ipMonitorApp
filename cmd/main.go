package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/AdemarTellecher/ipmonitorapp/internal/controller"
	"github.com/AdemarTellecher/ipmonitorapp/internal/model"
	"github.com/AdemarTellecher/ipmonitorapp/internal/view"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
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

	var listData []string
	selectedIndex := -1

	// Funções de manipulação para passar para a view
	refreshList := func() {}
	updateStatus := func() {}
	addIP := func() {}
	removeIP := func() {}
	onSelect := func(id int) {}

	// Inicializa a view
	ui := view.NewMainView(myApp, func() { updateStatus() }, func() { updateStatus() }, func() { addIP() }, func() { removeIP() }, func(id int) { onSelect(id) })

	// Redefine as funções agora que os componentes existem
	refreshList = func() {
		ips, _ := ctrl.ListIPs()
		listData = make([]string, len(ips))
		total, on, off := 0, 0, 0
		ui.IPList.Length = func() int { return len(listData) }
		ui.IPList.UpdateItem = func(i int, o fyne.CanvasObject) {
			row := o.(*fyne.Container)
			label := row.Objects[0].(*widget.Label)
			circle := row.Objects[2].(*canvas.Circle)
			label.SetText(listData[i])
			if len(ips) > i {
				if ips[i].Status == "Online" {
					circle.FillColor = color.RGBA{0, 200, 0, 255} // verde
				} else if ips[i].Status == "Offline" {
					circle.FillColor = color.RGBA{200, 0, 0, 255} // vermelho
				} else {
					circle.FillColor = color.RGBA{180, 180, 180, 255} // cinza
				}
				circle.Refresh()
			}
		}
		for i, d := range ips {
			listData[i] = d.IP + " [" + d.Status + "]"
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
		ui.IPList.Refresh()
	}

	updateStatus = func() {
		go func() {
			ctrl.UpdateAllStatuses()
			if drv, ok := fyne.CurrentApp().Driver().(interface{ CallOnMainThread(func()) }); ok {
				drv.CallOnMainThread(func() {
					refreshList()
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
		refreshList()
	}

	removeIP = func() {
		if selectedIndex >= 0 && selectedIndex < len(listData) {
			ip := extractIP(listData[selectedIndex])
			_ = ctrl.RemoveIP(ip)
			refreshList()
			selectedIndex = -1
		}
	}

	onSelect = func(id int) {
		selectedIndex = id
	}

	refreshList()

	// Atualização automática a cada 1 minuto
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			ctrl.UpdateAllStatuses()
			if drv, ok := fyne.CurrentApp().Driver().(interface{ CallOnMainThread(func()) }); ok {
				drv.CallOnMainThread(func() {
					refreshList()
				})
			}
		}
	}()

	ui.Window.ShowAndRun()
}

func extractIP(label string) string {
	for i := 0; i < len(label); i++ {
		if label[i] == ' ' {
			return label[:i]
		}
	}
	return label
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
