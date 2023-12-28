package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/playmixer/pc.assistent/pkg/player"
	"github.com/playmixer/pc.assistent/pkg/smarty"
	"github.com/playmixer/pc.assistent/pkg/yandex"

	"github.com/playmixer/pc.assistent/pkg/logger"
	voskclient "github.com/playmixer/pc.assistent/pkg/vosk-client"

	"github.com/joho/godotenv"
)

/**
 TODO
 - перезагрузка команд

**/

var (
	log       *logger.Logger
	assistent *smarty.Assiser
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

func LoadCommand() error {
	log.INFO("Loading command from store...")
	_cmd, err := Store.Open(Command{}).All()
	if err != nil {
		log.ERROR(err.Error())
		return err
	}
	assistent.DeleteAllCommand()
	for _, v := range _cmd.(map[string]Command) {
		log.INFO(fmt.Sprintf("\t upload command: %s", v.Commands[0]))
		assistent.AddGenCommand(smarty.ObjectCommand{
			Type:     smarty.TypeCommand(v.Type),
			Path:     v.Path,
			Args:     v.Args,
			Commands: v.Commands,
		})
	}
	return nil
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
	_, err = StoreInit(".storage")
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
	// speach := tts.New(tts.TTSProviderYandex)

	// Asisstent listener
	assistent = smarty.New(ctx)
	assistent.SetRecognizeCommand(recognizer)
	assistent.SetRecognizeName(recognizer)
	assistent.SetConfig(smarty.Config{
		Names:         []string{"альфа", "бета", "бэта"},
		ListenTimeout: time.Second * 1,
	})
	assistent.SetLogger(log)

	// Озвучка текста
	assistent.SetTTS(func(text string) error {
		ydx := yandex.New(os.Getenv("YANDEX_API_KEY"), os.Getenv("YANDEX_FOLDER_ID"))

		req := ydx.Speach(text)
		b, err := req.Post()
		if err != nil {
			return err
		}

		player.PlayMp3FromBytes(b)
		return nil
	})

	// Загрузка команд из хранилища
	LoadCommand()

	// Голосовые команды
	// assistent.AddCommand([]string{"который час", "сколько время"}, func(ctx context.Context, a *smarty.Assiser) {
	// 	txt := fmt.Sprint("Текущее время:", time.Now().Format("15:04"))
	// 	a.Print(txt)
	// 	err := a.Voice(txt)
	// 	if err != nil {
	// 		log.ERROR(err.Error())
	// 	}
	// 	// log.ERROR(a.Voice(txt).Error())
	// })
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
	httpPort := smarty.Getenv("HTTP_SERVER_PORT", "8080")
	server := HttpServerNew(httpPort)
	server.SetWSHandle(WSHandle)
	r := server.GetRoute()
	r.Use(CORSMiddleware())
	v0 := r.Group("/api/v0")
	{
		v0.GET("/command", httpGetCommands)
		v0.POST("/command", httpNewCommand)
		v0.DELETE("/command", httpDeleteCommand)
		v0.PUT("/command", httpUpdateCommand)
		v0.GET("/info", httpInfo)
		v0.POST("/refresh", httpRefresh)
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
				}
				if e == smarty.AEStartListeningName {
					log.INFO("Event:", "Listening name")
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
