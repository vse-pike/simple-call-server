package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"simplecall/server/internal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func translator(h *internal.Hub, w http.ResponseWriter, r *http.Request) {
	connection, _ := upgrader.Upgrade(w, r, nil)
	defer connection.Close()

	ch := make(chan internal.Envelope, 16)

	room := r.URL.Query().Get("room")
	name := r.URL.Query().Get("name")

	client := internal.NewClient(room, name, h, connection, ch)

	go client.WritePump()

	client.ReadPump()
}

func main() {
	hub := internal.NewHub()

	go hub.Run()

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Получен запрос на метод /healthz")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		translator(hub, w, r)
	})

	srv := &http.Server{Addr: ":8080"}

	// Слушаем в отдельной горутине -- ListenAndServe блокирующий,
	// а main() нужно освободить, чтобы дальше ждать сигнал остановки.
	go func() {
		slog.Info("Запуск сервера...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Ошибка запуска сервера", "Error", err)
		}
	}()

	// Блокируемся здесь, пока ОС не пришлёт сигнал на завершение
	// (Ctrl+C -> SIGINT, `docker stop`/systemd/k8s -> SIGTERM).
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	slog.Info("Получен сигнал остановки, завершаем работу...")

	// Даём серверу до 10 секунд на то, чтобы доработать уже открытые
	// запросы и корректно закрыть listener, прежде чем выйти принудительно.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Ошибка при graceful shutdown", "Error", err)
	} else {
		slog.Info("Сервер остановлен корректно")
	}
}
