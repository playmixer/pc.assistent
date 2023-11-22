package main

import (
	"context"
	"strconv"
	"time"

	"pc.assistent/internal/sayme"
	"pc.assistent/internal/smarty"

	"github.com/playmixer/pc.assistent/pkg/logger"

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

	assistent := smarty.New(ctx)
	assistent.SetLogger(log)

	assistent.AddCommand([]string{"который час", "сколько время"}, func(ctx context.Context, a *smarty.Assiser) {
		a.Print("Текущее время:", time.Now().Format(time.TimeOnly))
	})
	assistent.AddCommand([]string{"отключись", "выключись"}, func(ctx context.Context, a *smarty.Assiser) {
		a.Print("Отключаюсь")
		cancel()
	})

	go func() {
	waitEvent:
		for {
			select {
			case <-ctx.Done():
				log.DEBUG("Stop waitEvent")
				break waitEvent
			case e := <-assistent.GetSignalEvent():
				log.DEBUG("Event", strconv.Itoa(int(e)))
				if e == smarty.AEStartListening {
					log.INFO("Event:", "Listening command")
					sayme.New().TalkListen()
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
