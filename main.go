package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	smarthome "pc.assistent/internal/smart-home"
	socketserver "pc.assistent/internal/socket-server"
	"pc.assistent/internal/traymenu"
	"pc.assistent/internal/tts"

	"github.com/playmixer/pc.assistent/pkg/logger"
	voskclient "github.com/playmixer/pc.assistent/pkg/vosk-client"

	// "github.com/playmixer/pc.assistent/pkg/smarty"
	"pc.assistent/pkg/smarty"

	"github.com/joho/godotenv"
)

var (
	log *logger.Logger
)

type SocketSendEvent struct {
	Event int `json:"event"`
}

func MarshalSocketSendEvent(event int) []byte {

	res, _ := json.Marshal(SocketSendEvent{
		Event: event,
	})

	return res
}

func main() {
	ctx, cancel := context.WithCancel(context.Background()) //вся программа
	//логер
	log = logger.New("app")
	log.LogLevel = logger.DEBUG

	// Загрузка .env
	err := godotenv.Load()
	if err != nil {
		log.ERROR("Error loading .env file")
	}

	// Распознование речи
	recognizer := voskclient.New(log)

	// Озвучка текста
	speach := tts.New()

	// Иконка в трее
	traymenu.Init()

	// Asisstent listener
	assistent := smarty.New(ctx)
	assistent.SetRecognizeCommand(recognizer)
	assistent.SetRecognizeName(recognizer)
	assistent.SetConfig(smarty.Config{
		Names:                []string{"альфа", "бета", "бэта"},
		ListenNameTimeout:    time.Second * 3,
		ListenCommandTimeout: time.Second * 6,
	})
	assistent.SetLogger(log)
	assistent.SetTTS(speach.Voice)
	assistent.LoadCommands("./commands.json")

	// API smart home
	shome, err := smarthome.FactoryNew(smarthome.SHTuyaService)
	if err != nil {
		log.ERROR(err.Error())
	}

	// Голосовые команды
	assistent.AddCommand([]string{"который час", "сколько время"}, func(ctx context.Context, a *smarty.Assiser) {
		txt := fmt.Sprint("Текущее время:", time.Now().Format(time.TimeOnly))
		a.Print(txt)
		// log.ERROR(a.Voice(txt).Error())
	})
	assistent.AddCommand([]string{"отключись", "выключись"}, func(ctx context.Context, a *smarty.Assiser) {
		a.Print("Отключаюсь")
		// log.ERROR(a.Voice("Отключаюсь").Error())
		cancel()
	})
	assistent.AddCommand([]string{"включи свет в ванне", "включи свет в ванной"}, func(ctx context.Context, a *smarty.Assiser) {
		shome.PostDevice("787166238cce4e149625", shome.NewCommand().Add("switch_1", true))
	})
	assistent.AddCommand([]string{"выключи свет в ванне", "выключи свет в ванной"}, func(ctx context.Context, a *smarty.Assiser) {
		shome.PostDevice("787166238cce4e149625", shome.NewCommand().Add("switch_1", false))
	})
	assistent.AddCommand([]string{"включи свет в туалете"}, func(ctx context.Context, a *smarty.Assiser) {
		shome.PostDevice("787166238cce4e149625", shome.NewCommand().Add("switch_2", true))
	})
	assistent.AddCommand([]string{"выключи свет в туалете"}, func(ctx context.Context, a *smarty.Assiser) {
		shome.PostDevice("787166238cce4e149625", shome.NewCommand().Add("switch_2", false))
	})
	assistent.AddCommand([]string{"включи везде свет", "включи весь свет"}, func(ctx context.Context, a *smarty.Assiser) {
		shome.PostDevice("787166238cce4e149625", shome.NewCommand().Add("switch_1", true).Add("switch_2", true))
	})
	assistent.AddCommand([]string{"выключи везде свет", "выключи весь свет", "выключи свет везде"}, func(ctx context.Context, a *smarty.Assiser) {
		shome.PostDevice("787166238cce4e149625", shome.NewCommand().Add("switch_1", false).Add("switch_2", false))
	})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// Сокет сервер
	socket := socketserver.StartServer(func(message []byte) {
		fmt.Println(string(message))
	})

	// События асистента
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
				log.DEBUG("Event", fmt.Sprint(e))
				if e == smarty.AEStartListening {
					log.INFO("Event:", "Listening command")
					// sayme.New().SoundStart()
					traymenu.SetIcon(".\\command.ico")
				}
				if e == smarty.AEStopListening {
					log.INFO("Event:", "Stop listening command")
					// sayme.New().SoundEnd()
					traymenu.SetIconDefault()
				}
				if e == smarty.AEApplyCommand {
					log.INFO("Event:", "Run command")
				}

				socket.WriteMessage(MarshalSocketSendEvent(int(e)))
			}
		}
	}()

	log.INFO("Starting App")
	assistent.Start()

	log.INFO("Stoping App...")
	time.Sleep(time.Second * 1)
	log.INFO("Stop App")
}
