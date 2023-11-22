package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"pc.assistent/internal/sayme"

	"github.com/playmixer/pc.assistent/pkg/logger"
	"github.com/playmixer/pc.assistent/pkg/smarty"
	voskclient "github.com/playmixer/pc.assistent/pkg/vosk-client"

	"github.com/joho/godotenv"
)

var (
	log *logger.Logger
)

func main() {
	ctx, cancel := context.WithCancel(context.Background()) //вся программа
	//логер
	log = logger.New("app")
	log.LogLevel = logger.INFO

	err := godotenv.Load()
	if err != nil {
		log.ERROR("Error loading .env file")
	}

	recognizer := voskclient.New(log)

	assistent := smarty.New(ctx, recognizer)
	assistent.SetConfig(smarty.Config{
		Names:                []string{"альфа", "бета", "бэта"},
		ListenNameTimeout:    time.Second * 3,
		ListenCommandTimeout: time.Second * 6,
	})
	assistent.SetLogger(log)

	assistent.AddCommand([]string{"который час", "сколько время"}, func(ctx context.Context, a *smarty.Assiser) {
		a.Print("Текущее время:", time.Now().Format(time.TimeOnly))
	})
	assistent.AddCommand([]string{"отключись", "выключись"}, func(ctx context.Context, a *smarty.Assiser) {
		a.Print("Отключаюсь")
		cancel()
	})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	go func() {
	waitEvent:
		for {
			select {
			case <-sigs:
				break waitEvent

			case <-ctx.Done():
				log.DEBUG("Stop waitEvent")
				break waitEvent

			case e := <-assistent.GetSignalEvent():
				log.DEBUG("Event", strconv.Itoa(int(e)))
				if e == smarty.AEStartListening {
					log.INFO("Event:", "Listening command")
					sayme.New().SoundStart()
				}
				if e == smarty.AEStopListening {
					log.INFO("Event:", "Stop listening command")
					sayme.New().SoundEnd()
				}
			}
		}
	}()

	log.INFO("Starting App")
	assistent.Start()

	log.INFO("Stoping App...")
	time.Sleep(time.Second * 1)
	log.INFO("Stop App")
}
