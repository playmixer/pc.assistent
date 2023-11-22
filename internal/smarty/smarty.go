package smarty

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"pc.assistent/pkg/listen"

	voskclient "github.com/playmixer/pc.assistent/pkg/vosk-client"

	ls "github.com/texttheater/golang-levenshtein/levenshtein"
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

func inCmd(cmd string, cArray []string) bool {
	for _, c := range cArray {
		if d := cmdDistance(c, cmd); d == 0 {
			fmt.Println(c, cmd, "distance", d)
			return true
		}
	}
	return false
}

func cmdDistance(t1, t2 string) int {
	return ls.DistanceForStrings([]rune(t1), []rune(t2), ls.DefaultOptions)
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

type Assiser struct {
	ctx                  context.Context
	log                  iLogger
	Names                []string
	ListenNameTimeout    time.Duration
	ListenCommandTimeout time.Duration
	commands             []CommandStruct
	eventChan            chan AssiserEvent
}

func New(ctx context.Context) *Assiser {

	a := &Assiser{
		log:                  &logger{},
		ctx:                  ctx,
		Names:                []string{"альфа", "бета", "бэта"},
		ListenNameTimeout:    time.Second * 2,
		ListenCommandTimeout: time.Second * 6,
		commands:             make([]CommandStruct, 0),
		eventChan:            make(chan AssiserEvent, 1),
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
	for i, command := range a.commands {
		if inCmd(cmd, a.commands[i].Commands) {
			a.log.INFO("Run command", cmd)
			ctx, cancel := context.WithCancel(context.Background())
			a.commands[i].Context = ctx
			a.commands[i].Cancel = cancel
			go func() {
				a.commands[i].IsActive = true
				command.Func(ctx, a)
				a.commands[i].IsActive = false
				a.commands[i].Cancel()
			}()
			break
		}
	}

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

	// слушает только когда назвали имя 1ый поток
	log.INFO("Initilize listen command. Stram #1 ")
	rCom1 := listen.New(a.ListenCommandTimeout)
	rCom1.SetLogger(log)
	defer rCom1.Stop()

	// слушает только когда назвали имя 2ой поток
	log.INFO("Initilize listen command. Stram #2 ")
	rCom2 := listen.New(a.ListenCommandTimeout)
	rCom2.SetLogger(log)
	defer rCom2.Stop()

	recognize := voskclient.New(log)

	var lastMessage = listenMessage{
		t: time.Now(),
	}

	listenName := any2ChanResult(ctx, r1.WavCh, r2.WavCh)
	listenCommand := any2ChanResult(ctx, rCom1.WavCh, rCom2.WavCh)

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
			txt, err := recognize.Recognize(s)
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
							rCom1.Start(ctx)
							go func() {
								time.Sleep(a.ListenCommandTimeout / 2)
								rCom2.Start(ctx)
							}()
						}
						break
					}
				}
				lastMessage.SetMessage(txt)
			}

		//слушаем команду
		case com := <-listenCommand:
			txt, err := recognize.Recognize(com)
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
				if rCom1.IsActive || rCom2.IsActive {
					if time.Since(lastMessage.t) > a.ListenCommandTimeout {
						a.PostSiglanEvent(AEStopListening)
						log.DEBUG("Stop listen command")
						rCom1.Stop()
						rCom2.Stop()
						a.PostSiglanEvent(AEStartListeningName)
						r1.Start(ctx)
						time.Sleep(a.ListenNameTimeout)
						r2.Start(ctx)
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
