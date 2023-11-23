package smarty

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	fuzzy "github.com/paul-mannino/go-fuzzywuzzy"
	"github.com/playmixer/pc.assistent/pkg/listen"
	"golang.org/x/exp/slices"
)

type AssiserEvent int

const (
	AEStartListening     AssiserEvent = 10
	AEStopListening      AssiserEvent = 20
	AEStartListeningName AssiserEvent = 30
	AEStopSListeningName AssiserEvent = 40
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
	commands             []CommandStruct
	eventChan            chan AssiserEvent
	recognize            IReacognize
}

func New(ctx context.Context, recognize IReacognize) *Assiser {

	a := &Assiser{
		log:                  &logger{},
		ctx:                  ctx,
		Names:                []string{"альфа", "alpha"},
		ListenNameTimeout:    time.Second * 2,
		ListenCommandTimeout: time.Second * 6,
		commands:             make([]CommandStruct, 0),
		eventChan:            make(chan AssiserEvent, 1),
		recognize:            recognize,
	}

	return a
}

func (a *Assiser) AddCommand(cmd []string, f CommandFunc) {
	a.commands = append(a.commands, CommandStruct{
		Commands: cmd,
		Func:     f,
	})
}

func (a *Assiser) runCommand(cmd string) {
	i, percent := a.RotateCommand(cmd)
	a.log.INFO("rotate command", cmd, fmt.Sprint(i), fmt.Sprint(percent))
	if percent == 100 {
		a.log.INFO("Run command", cmd)
		ctx, cancel := context.WithCancel(context.Background())
		a.commands[i].Context = ctx
		a.commands[i].Cancel = cancel
		go func() {
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

	ctx, cancel := context.WithCancel(context.Background()) //вся программа

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// слушаем имя 1ый поток
	log.INFO("Starting listen. Stram #1 ")
	r1 := listen.New(a.ListenNameTimeout)
	r1.SetLogger(log)
	r1.Start(ctx)
	defer r1.Stop()

	r2 := listen.New(a.ListenNameTimeout)
	go func() {
		//задержка перед вторым потоком что бы слушать их асинхронно
		time.Sleep(a.ListenNameTimeout / 2)
		log.INFO("Starting listen. Stram #2 ")
		r2.SetLogger(log)
		r2.Start(ctx)
	}()
	// слушаем имя 2ой поток
	defer r2.Stop()

	// слушает только команду 1ый поток
	log.INFO("Initilize listen command. Stram #1 ")
	rCheckCommand := listen.New(time.Second)
	rCheckCommand.SetLogger(log)
	defer rCheckCommand.Stop()

	// слушает только команду 2ой поток
	log.INFO("Initilize listen command. Stram #2 ")
	rCommand := listen.New(time.Minute)
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
		select {
		case <-sigs:
			cancel()
			break waitFor
		case <-a.ctx.Done():
			cancel()
			break waitFor
		//слушаем имя
		case s := <-listenName:
			txt, err := a.recognize.Recognize(s)
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
						r1.Stop()
						r2.Stop()
						//начинаем слушать голосовую команду
						if !lastMessage.IsActualMessage(txt, a.ListenNameTimeout) {
							log.DEBUG("Start listen command")
							a.PostSiglanEvent(AEStartListening)
							rCheckCommand.Start(ctx)
							rCommand.Start(ctx)
						}
						break
					}
				}
				lastMessage.SetMessage(txt)
			}

		//проверяем есть ли продолжение команды
		case ct := <-rCheckCommand.WavCh:
			txt, err := a.recognize.Recognize(ct)
			if err != nil {
				log.ERROR(err.Error())
			}

			log.DEBUG("check command: ", txt)
			//проверяем было ли что то сказано
			if txt == "" {
				rCommand.SliceRecod()
			}

		//слушаем команду
		case com := <-rCommand.WavCh:
			txt, err := a.recognize.Recognize(com)
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
				if rCheckCommand.IsActive || rCommand.IsActive {
					if time.Since(lastMessage.t) > a.ListenCommandTimeout {
						a.PostSiglanEvent(AEStopListening)
						log.DEBUG("Stop listen command")
						rCheckCommand.Stop()
						rCommand.Stop()
						a.PostSiglanEvent(AEStartListeningName)
						go func() {
							r1.Start(ctx)
							time.Sleep(a.ListenNameTimeout / 2)
							r2.Start(ctx)
						}()
					}
				}
			}

		//слушаем событие
		case e := <-a.GetSignalEvent():
			log.INFO(strconv.Itoa(int(e)))
		}
	}
}

func (a *Assiser) Print(t ...any) {
	fmt.Println(a.Names[0]+": ", fmt.Sprint(t...))
}

func (a *Assiser) PostSiglanEvent(s AssiserEvent) {
	a.eventChan <- s
}

func (a *Assiser) GetSignalEvent() <-chan AssiserEvent {

	return a.eventChan
}
