package main

import (
	"embed"
	"log"
	"time"

	"github.com/AdemarTellecher/ipmonitorapp/internal/model"
	"github.com/AdemarTellecher/ipmonitorapp/internal/service"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	// 1. Resolve dinamicamente o caminho do banco SQLite na pasta do executável
	dbPath := model.ResolveDBPath("ipmonitorDB.db")
	repo, err := model.NewRepository(dbPath)
	if err != nil {
		log.Fatalf("Erro ao inicializar o banco SQLite portátil: %v", err)
	}
	defer repo.Close()

	// 2. Cria o serviço de monitoramento exposto para o frontend
	monitorService := service.NewMonitorService(repo)

	// 3. Inicializa o aplicativo Wails v3
	app := application.New(application.Options{
		Name:        "IP Monitor App",
		Description: "Monitoramento de conectividade ICMP Ping portátil e ultra-leve",
		Services: []application.Service{
			application.NewServiceWithOptions(monitorService, application.ServiceOptions{
				Route: "/api",
			}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	// 4. Cria a janela nativa elegante
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "IP Monitor",
		Width:            540,
		Height:           740,
		MinWidth:         480,
		MinHeight:        640,
		InitialPosition:  application.WindowCentered,
		BackgroundColour: application.NewRGB(16, 19, 26),
		URL:              "/",
	})

	// 5. Rotina periódica de ping a cada 1 minuto com emissão de evento
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			updated, err := monitorService.UpdateAllStatuses()
			if err == nil {
				app.Event.Emit("ips-updated", updated)
			}
		}
	}()

	// 6. Executa a aplicação
	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
