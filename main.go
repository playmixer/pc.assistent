package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/playmixer/pc.assistent/internal/api"
	"github.com/playmixer/pc.assistent/internal/httpserver"
	smarthome "github.com/playmixer/pc.assistent/internal/smart-home"
	"github.com/playmixer/pc.assistent/internal/store"
	"github.com/playmixer/pc.assistent/internal/traymenu"
	"github.com/playmixer/pc.assistent/internal/tts"
	"github.com/playmixer/pc.assistent/pkg/smarty"

	"github.com/playmixer/pc.assistent/pkg/logger"
	voskclient "github.com/playmixer/pc.assistent/pkg/vosk-client"

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
	log.LogLevel = logger.INFO

	log.INFO("Starting App")
	// Загрузка .env
	err := godotenv.Load()
	if err != nil {
		log.ERROR("Error loading .env file")
	}

	//store
	db, err := store.Init("data.ojs")
	if err != nil {
		log.ERROR(err.Error())
		panic(err)
	}

	// Распознование речи
	recognizer := voskclient.New()
	rLog := logger.New("recognize")
	rLog.LogLevel = logger.INFO
	recognizer.SetLogger(rLog)

	// Озвучка текста
	speach := tts.New()

	// Иконка в трее
	traymenu.Init()

	// Asisstent listener
	assistent := smarty.New(ctx)
	assistent.SetRecognizeCommand(recognizer)
	assistent.SetRecognizeName(recognizer)
	assistent.SetConfig(smarty.Config{
		Names:         []string{"альфа", "бета", "бэта"},
		ListenTimeout: time.Second * 1,
	})
	assistent.SetLogger(log)
	assistent.SetTTS(speach.Voice)
	// assistent.LoadCommands("./commands.json")

	// API smart home
	sHome, err := smarthome.FactoryNew(smarthome.SHTuyaService)
	if err != nil {
		log.ERROR(err.Error())
	}

	// Загрузка команд из хранилища
	log.INFO("Loading command from store...")
	_cmd, err := db.Open(store.Command{}).All()
	if err != nil {
		log.ERROR(err.Error())
	}
	for _, v := range _cmd.(map[string]store.Command) {
		log.INFO(fmt.Sprintf("\t upload command: %s", v.Commands[0]))
		assistent.AddGenCommand(smarty.ObjectCommand{
			Type:     smarty.TypeCommand(v.Type),
			Path:     v.Path,
			Args:     v.Args,
			Commands: v.Commands,
		})
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
	assistent.AddCommand([]string{"счет", "счёт"}, func(ctx context.Context, a *smarty.Assiser) {
		ticker := time.NewTicker(time.Second)
		counter := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				counter += 1
				log.INFO(fmt.Sprintf("counter %v", counter))
			}
		}
	})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// Сокет сервер
	httpPort := Getenv("HTTP_SERVER_PORT", "8080")
	server := httpserver.New(httpPort)
	server.SetWSHandle(api.WSHandle)
	r := server.GetRoute()
	v0 := r.Group("/api/v0")
	{
		v0.GET("/command", api.GetCommands)
		v0.POST("/command", api.NewCommand)
		v0.DELETE("/command", api.DeleteCommand)
		v0.PUT("/command", api.UpdateCommand)
	}

	server.Start()
	log.INFO("Start web server :" + httpPort)

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
				if e == smarty.AEStartListeningCommand {
					log.INFO("Event:", "Listening command")
					// sayme.New().SoundStart()
					traymenu.SetIcon(".\\command.ico")
				}
				if e == smarty.AEStartListeningName {
					log.INFO("Event:", "Listening name")
					// sayme.New().SoundEnd()
					traymenu.SetIconDefault()
				}
				if e == smarty.AEApplyCommand {
					log.INFO("Event:", "Run command")
				}

				server.WriteMessage(MarshalSocketSendEvent(int(e)))
			}
		}
	}()

	log.INFO("Starting Listener")
	assistent.Start()

	log.INFO("Stoping App...")
	time.Sleep(time.Second * 3)
	log.INFO("Stop App")
}

func Getenv(key string, def string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}

	return def
}
