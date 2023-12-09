package smarty

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	fuzzy "github.com/paul-mannino/go-fuzzywuzzy"
	// "github.com/playmixer/pc.assistent/pkg/listen"
	"golang.org/x/exp/slices"
	"pc.assistent/pkg/listen"
)

type AssiserEvent int

const (
	AEStartListening     AssiserEvent = 10
	AEStopListening      AssiserEvent = 20
	AEStartListeningName AssiserEvent = 30
	AEStopSListeningName AssiserEvent = 40
	AEApplyCommand       AssiserEvent = 50

	F_MIN_TOKEN    = 90
	F_MAX_DISTANCE = 10
	F_MIN_RATIO    = 75
)

type iLogger interface {
	ERROR(v ...string)
	INFO(v ...string)
	DEBUG(v ...string)
}

type logger struct{}

func (l *logger) ERROR(v ...string) {
	log.Println("ERROR", v)
}

func (l *logger) INFO(v ...string) {
	log.Println("INFO", v)
}

func (l *logger) DEBUG(v ...string) {
	log.Println("DEBUG", v)
}

type listenMessage struct {
	_message string
	t        time.Time
}

func (lm *listenMessage) SetMessage(m string) {
	lm.t = time.Now()
	lm._message = m
}

func (lm *listenMessage) IsActualMessage(m string, d time.Duration) bool {
	if m == lm._message && time.Since(lm.t) < d {
		return true
	}

	return false
}

func any2ChanResult(ctx context.Context, c1 chan []byte, c2 chan []byte) chan []byte {
	result := make(chan []byte, 1)
	go func() {
	waitChan2:
		for {
			select {
			case <-ctx.Done():
				break waitChan2
			case r := <-c1:
				result <- r
			case r := <-c2:
				result <- r
			}

		}
	}()
	return result
}

type CommandFunc func(ctx context.Context, a *Assiser)

type CommandStruct struct {
	Commands []string
	Func     CommandFunc
	IsActive bool
	Context  context.Context
	Cancel   context.CancelFunc
}

type Config struct {
	Names                []string
	ListenNameTimeout    time.Duration
	ListenCommandTimeout time.Duration
}

type IReacognize interface {
	Recognize(bufWav []byte) (string, error)
}

type Assiser struct {
	ctx                  context.Context
	log                  iLogger
	Names                []string
	ListenNameTimeout    time.Duration
	ListenCommandTimeout time.Duration
	isListenName         bool
	commands             []CommandStruct
	eventChan            chan AssiserEvent
	recognizeCommand     IReacognize
	recognizeName        IReacognize
	VoiceEnable          bool
	voiceFunc            func(text string) error
	sync.Mutex
}

func New(ctx context.Context) *Assiser {

	a := &Assiser{
		log:                  &logger{},
		ctx:                  ctx,
		Names:                []string{"альфа", "alpha"},
		ListenNameTimeout:    time.Second * 2,
		ListenCommandTimeout: time.Second * 6,
		isListenName:         true,
		commands:             make([]CommandStruct, 0),
		eventChan:            make(chan AssiserEvent, 1),
		VoiceEnable:          true,
		voiceFunc:            func(text string) error { return nil },
	}

	return a
}

func (a *Assiser) SetRecognizeCommand(recognize IReacognize) {
	a.recognizeCommand = recognize
}

func (a *Assiser) SetRecognizeName(recognize IReacognize) {
	a.recognizeName = recognize
}

func (a *Assiser) AddCommand(cmd []string, f CommandFunc) {
	a.commands = append(a.commands, CommandStruct{
		Commands: cmd,
		Func:     f,
	})
}

func (a *Assiser) runCommand(cmd string) {
	i, found := a.RotateCommand2(cmd)
	a.log.DEBUG("rotate command", cmd, fmt.Sprint(i), fmt.Sprint(found))
	if found {
		a.log.DEBUG("Run command", cmd)
		a.PostSiglanEvent(AEApplyCommand)
		ctx, cancel := context.WithCancel(context.Background())
		a.commands[i].Context = ctx
		a.commands[i].Cancel = cancel
		go func() {
			a.Lock()
			defer a.Unlock()
			a.commands[i].IsActive = true
			a.commands[i].Func(ctx, a)
			a.commands[i].IsActive = false
			a.commands[i].Cancel()
		}()
	}

}

func (a *Assiser) RotateCommand(talk string) (index int, percent int) {
	var idx int = 0
	percent = 0
	var founded bool = false
	for i, command := range a.commands {
		for _, c := range command.Commands {
			//проверяем что все слова команды есть в предложении пользователя
			wordsCommand := strings.Fields(c)
			wordsTalk := strings.Fields(talk)
			allWordsInCommand := true
			for _, word := range wordsCommand {
				if ok := slices.Contains(wordsTalk, word); !ok {
					allWordsInCommand = false
					break
				}
			}
			p := fuzzy.TokenSetRatio(c, talk)
			if p > percent && allWordsInCommand {
				idx, percent = i, p
				founded = true
			}
		}
	}
	if !founded {
		return 0, 0
	}

	return idx, percent
}

func (a *Assiser) FoundCommandByToken(talk string) (index, percent int, founded bool) {
	var idx int = 0
	percent = 0
	founded = false
	for i, command := range a.commands {
		for _, c := range command.Commands {
			p := fuzzy.TokenSetRatio(c, talk)
			if p > percent {
				idx, percent = i, p
				founded = true
			}
			// fmt.Printf("%s vs %s, percent=%v, founded=%v\n", talk, c, p, founded)
		}
	}
	return idx, percent, founded
}

func (a *Assiser) FoundCommandByDistance(talk string) (index, distance int, founded bool) {
	var idx int = 0
	distance = 1000000
	founded = false
	for i, command := range a.commands {
		for _, c := range command.Commands {
			d := fuzzy.EditDistance(c, talk)
			if d < distance {
				idx, distance = i, d
				founded = true
			}
			// fmt.Printf("%s vs %s, distance=%v, founded=%v\n", talk, c, d, founded)
		}
	}
	return idx, distance, founded
}

func (a *Assiser) FoundCommandByRatio(talk string) (index, ratio int, founded bool) {
	var idx int = 0
	ratio = 0
	founded = false
	for i, command := range a.commands {
		for _, c := range command.Commands {
			r := fuzzy.Ratio(c, talk)
			if r > ratio {
				idx, ratio = i, r
				founded = true
			}
			// fmt.Printf("%s vs %s, ratio=%v, founded=%v\n", talk, c, r, founded)
		}
	}
	return idx, ratio, founded
}

func (a *Assiser) RotateCommand2(talk string) (index int, found bool) {
	var idx int = 0
	found = false
	t, tv, tf := a.FoundCommandByToken(talk)
	d, dv, df := a.FoundCommandByDistance(talk)
	r, rv, rf := a.FoundCommandByRatio(talk)
	if tf == df && df == rf && rf &&
		t == d && d == r &&
		tv >= F_MIN_TOKEN && (dv <= F_MAX_DISTANCE || rv >= F_MIN_RATIO) {
		return t, tf
	}

	return idx, false
}

func (a *Assiser) SetConfig(cfg Config) {
	a.Names = cfg.Names
	a.ListenNameTimeout = cfg.ListenNameTimeout
	a.ListenCommandTimeout = cfg.ListenCommandTimeout
}

func (a *Assiser) SetLogger(log iLogger) {
	a.log = log
}

func (a *Assiser) Start() {
	log := a.log
	if a.recognizeCommand == nil || a.recognizeName == nil {
		log.ERROR("Cannot founded recognize method")
		return
	}

	ctx, cancel := context.WithCancel(context.Background()) //вся программа

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// слушаем имя 1ый поток
	log.INFO("Starting listen. Stram #1 ")
	r1 := listen.New(a.ListenNameTimeout)
	r1.SetName("NameL1")
	r1.SetLogger(log)
	r1.Start(ctx)
	defer r1.Stop()

	r2 := listen.New(a.ListenNameTimeout)
	r2.SetName("NameL2")
	r2.SetLogger(log)
	go func() {
		//задержка перед вторым потоком что бы слушать их асинхронно
		time.Sleep(a.ListenNameTimeout / 2)
		log.INFO("Starting listen. Stram #2 ")
		r2.Start(ctx)
	}()
	// слушаем имя 2ой поток
	defer r2.Stop()

	// слушает паузы
	log.INFO("Initilize listen command. Stram #1 ")
	rCheckPause := listen.New(time.Second)
	rCheckPause.SetName("PauseL")
	rCheckPause.SetLogger(log)
	defer rCheckPause.Stop()

	// слушает только команду
	log.INFO("Initilize listen command. Stram #2 ")
	rCommand := listen.New(time.Minute)
	rCommand.SetName("CommandL")
	rCommand.SetLogger(log)
	defer rCommand.Stop()

	var lastMessage = listenMessage{
		t: time.Now(),
	}

	listenName := any2ChanResult(ctx, r1.WavCh, r2.WavCh)
	// listenCommand := any2ChanResult(ctx, rCom1.WavCh, rCom2.WavCh)

	a.AddCommand([]string{"стоп", "stop"}, func(ctx context.Context, a *Assiser) {
		for i := range a.commands {
			if a.commands[i].IsActive {
				log.DEBUG("Стоп", fmt.Sprint(a.commands[i]))
				a.commands[i].Cancel()
			}
		}
	})

waitFor:
	for {
		log.DEBUG("empty for loop")
		select {
		case <-sigs:
			cancel()
			break waitFor
		case <-a.ctx.Done():
			cancel()
			break waitFor
		//слушаем имя
		case s := <-listenName:
			// if !a.isListenName {
			// 	continue
			// }
			log.DEBUG("smarty" + " listening name ")
			txt, err := a.recognizeName.Recognize(s)
			if err != nil {
				log.ERROR(err.Error())
			}
			if txt != "" {
				if !lastMessage.IsActualMessage(txt, a.ListenNameTimeout) {
					log.INFO("Вы назвали имя:", txt)
				}
				for _, name := range a.Names {
					if txt == name {
						//стопаем слушать имя
						a.isListenName = false
						r1.Stop()
						r2.Stop()
						//начинаем слушать голосовую команду
						if !lastMessage.IsActualMessage(txt, a.ListenNameTimeout) {
							log.DEBUG("Start listen command")
							a.PostSiglanEvent(AEStartListening)
							rCheckPause.Start(ctx)
							rCommand.Start(ctx)
						}
						break
					}
				}
				lastMessage.SetMessage(txt)
			}

		//проверяем есть ли продолжение команды
		case ct := <-rCheckPause.WavCh:
			// if a.isListenName {
			// 	continue
			// }
			log.DEBUG("Readed pause from wav chanel")
			var err error
			txt := ""
			txt, err = a.recognizeName.Recognize(ct)
			if err != nil {
				log.ERROR(err.Error())
			}
			//проверяем было ли что то сказано
			if txt == "" {
				rCommand.SliceRecod()
			}

		//слушаем команду
		case com := <-rCommand.WavCh:
			// if a.isListenName {
			// 	continue
			// }
			log.DEBUG("Readed command from wav chanel")
			txt, err := a.recognizeCommand.Recognize(com)
			if err != nil {
				log.ERROR(err.Error())
			}
			if txt != "" {
				if !lastMessage.IsActualMessage(txt, a.ListenCommandTimeout) {
					log.INFO("Вы сказали:", txt)
					a.runCommand(txt)
				}
				lastMessage.SetMessage(txt)
			} else {
				if rCheckPause.IsActive || rCommand.IsActive {
					if time.Since(lastMessage.t) > a.ListenCommandTimeout {
						a.PostSiglanEvent(AEStopListening)
						log.DEBUG("Stop listen command")
						rCheckPause.Stop()
						rCommand.Stop()
						a.PostSiglanEvent(AEStartListeningName)
						a.isListenName = true
						// go func() {
						r1.Start(ctx)
						time.Sleep(a.ListenNameTimeout / 2)
						r2.Start(ctx)
						// }()
					}
				}
			}
		}
	}
}

func (a *Assiser) Print(t ...any) {
	fmt.Println(a.Names[0]+": ", fmt.Sprint(t...))
}

func (a *Assiser) PostSiglanEvent(s AssiserEvent) {
	select {
	case a.eventChan <- s:
	default:
	}

}

func (a *Assiser) GetSignalEvent() <-chan AssiserEvent {
	return a.eventChan
}

func (a *Assiser) SetTTS(f func(text string) error) {
	a.voiceFunc = f
}

func (a *Assiser) Voice(text string) error {
	if !a.VoiceEnable {
		return nil
	}

	return a.voiceFunc(text)
}

type typeCommand string

const (
	tcExec typeCommand = "exec"
)

type JsonCommand struct {
	Type     typeCommand `json:"type"`
	Path     string      `json:"path"`
	Args     []string    `json:"args"`
	Commands []string    `json:"commands"`
}

/**
* создаем команду для запуска внешнего процесса
 */
func (a *Assiser) newCommandExec(pathFile string, args ...string) CommandFunc {

	return func(ctx context.Context, a *Assiser) {
		go func() {
			exec.Command(pathFile, args...).Run()
		}()
	}
}

func (a *Assiser) LoadCommands(filepath string) error {
	cByte, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	var data []JsonCommand
	err = json.Unmarshal(cByte, &data)
	if err != nil {
		return err
	}

	for i, _ := range data {
		var f CommandFunc
		if data[i].Type == tcExec {
			f = a.newCommandExec(data[i].Path, data[i].Args...)
		}
		a.AddCommand(data[i].Commands, f)
	}
	return nil
}
