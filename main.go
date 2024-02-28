package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/playmixer/corvid/player"
	"github.com/playmixer/corvid/smarty"
	"github.com/playmixer/num2words"
	"github.com/playmixer/pc.assistent/functions"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/playmixer/corvid/logger"
	voskclient "github.com/playmixer/corvid/vosk-client"

	"github.com/joho/godotenv"
)

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
	_cmd, err := Store.Open(Command{}).All()
	if err != nil {
		log.ERROR(err.Error())
		return err
	}
	assistent.DeleteAllCommand()
	assistent.InitDefaultCommand()
	for _, v := range _cmd.(map[string]Command) {
		log.DEBUG(fmt.Sprintf("\t upload command: %s", v.Commands[0]))
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
	// recognizer2 := .RecognizeByte()

	rLog := logger.New("recognize")
	rLog.LogLevel = logger.INFO
	recognizer.SetLogger(rLog)

	// Asisstent listener
	assistent = smarty.New(ctx)
	assistent.SetRecognizeCommand(recognizer)
	assistent.SetRecognizeName(recognizer)
	assistent.SetConfig(smarty.Config{
		Names:         []string{"альфа"},
		ListenTimeout: time.Second * 1,
	})
	assistent.SetLogger(log)

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_API_KEY"))
	if err != nil {
		log.ERROR(err.Error())
		return
	}
	_chatId := os.Getenv("TELEGRAM_CHAT_ID")
	chatId, err := strconv.Atoi(_chatId)

	// Озвучка текста
	assistent.SetTTS(func(ctx context.Context, text string) error {
		//транслируем ответ в бота
		_, err = bot.Send(tgbotapi.NewMessage(int64(chatId), fmt.Sprintf("assistent: %s", text)))
		if err != nil {
			log.ERROR(err.Error())
		}

		text = functions.CleanString(text, " ")
		_text_splited := strings.Split(text, " ")
		for i, _t := range _text_splited {
			if functions.IsInt(_t) {
				_text_splited[i] = num2words.Convert(functions.StrToInt(_t))
			}
		}

		text = strings.Join(_text_splited, " ")

		var err error
		var b []byte
		requestURL := fmt.Sprintf("http://%s:%s/tts_to_wav?text=%s", smarty.Getenv("TTS_HOST", "localhost"), smarty.Getenv("TTS_PORT", "8000"), url.QueryEscape(text))
		res, err := http.Get(requestURL)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		b, err = io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		player.PlayWavFromBytes(ctx, b)
		return nil
	})

	// Загрузка команд из хранилища
	log.INFO("Loading command from store...")
	LoadCommand()

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
				cancel()
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

			case said := <-assistent.UserSaid:
				_, err = bot.Send(tgbotapi.NewMessage(int64(chatId), fmt.Sprintf("Вы: %s", said)))
				if err != nil {
					log.ERROR(err.Error())
				}
			}
		}
	}()

	log.INFO("Starting Listener")
	assistent.Start()

	log.INFO("Stoping App...")
	time.Sleep(time.Second * 1)
	log.INFO("Stop App")
}
