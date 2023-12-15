package smarty

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/playmixer/pc.assistent/pkg/listen"
	"golang.org/x/exp/slices"
)

type AssiserEvent int

const (
	AEStartListeningCommand AssiserEvent = 10
	AEStartListeningName    AssiserEvent = 20
	AEApplyCommand          AssiserEvent = 30

	F_MIN_TOKEN    = 90
	F_MAX_DISTANCE = 10
	F_MIN_RATIO    = 75

	LEN_WAV_BUFF = 20

	CMD_MAX_EMPTY_MESSAGE = 10
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
	Names         []string
	ListenTimeout time.Duration
}

type IReacognize interface {
	Recognize(bufWav []byte) (string, error)
}

type wavBuffer struct {
	buf  *[]byte
	text string
}

type Assiser struct {
	ctx              context.Context
	log              iLogger
	Names            []string
	ListenTimeout    time.Duration
	commands         []CommandStruct
	eventChan        chan AssiserEvent
	recognizeCommand IReacognize
	recognizeName    IReacognize
	VoiceEnable      bool
	voiceFunc        func(text string) error
	wavBuffer        []wavBuffer
	sync.Mutex
}

func New(ctx context.Context) *Assiser {

	a := &Assiser{
		log:           &logger{},
		ctx:           ctx,
		Names:         []string{"альфа", "alpha"},
		ListenTimeout: time.Second * 2,
		commands:      make([]CommandStruct, 0),
		eventChan:     make(chan AssiserEvent, 1),
		VoiceEnable:   true,
		voiceFunc:     func(text string) error { return nil },
		wavBuffer:     make([]wavBuffer, LEN_WAV_BUFF),
	}

	return a
}

func (a *Assiser) AddWavToBuf(w *wavBuffer) *[]wavBuffer {
	a.Lock()
	defer a.Unlock()
	a.wavBuffer = append([]wavBuffer{*w}, a.wavBuffer[:LEN_WAV_BUFF-1]...)

	return &a.wavBuffer
}

func (a *Assiser) GetWavFromBuf(count int) []byte {
	a.Lock()
	defer a.Unlock()
	result := *a.wavBuffer[0].buf
	for i := 1; i <= count && i <= LEN_WAV_BUFF && i < len(a.wavBuffer); i++ {
		if a.wavBuffer[i].buf != nil {
			result = listen.ConcatWav(*a.wavBuffer[i].buf, result)
		}
	}
	return result
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
	a.Lock()
	defer a.Unlock()
	cmd = CleanCommandFromName(a.Names, cmd)
	i, found := a.RotateCommand2(cmd)
	a.log.DEBUG("rotate command", cmd, fmt.Sprint(i), fmt.Sprint(found))
	if found {
		a.log.DEBUG("Run command", cmd)
		a.PostSignalEvent(AEApplyCommand)
		ctx, cancel := context.WithCancel(a.ctx)
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
	a.ListenTimeout = cfg.ListenTimeout
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

	ctx, cancel := context.WithCancel(a.ctx) //вся программа

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// слушаем имя 1ый поток
	log.INFO("Starting listen. Stram #1 ")
	record := listen.New(a.ListenTimeout)
	record.SetName("Record")
	record.SetLogger(log)
	record.Start(ctx)
	defer record.Stop()

	a.AddCommand([]string{"стоп", "stop"}, func(ctx context.Context, a *Assiser) {
		for i := range a.commands {
			if a.commands[i].IsActive {
				log.INFO("Стоп", fmt.Sprint(a.commands[i].Commands[0]))
				a.commands[i].Cancel()
			}
		}
	})

	var notEmptyMessageCounter = 0
	var emptyMessageCounter = 0
	var isListenName = true

waitFor:
	for {
		log.DEBUG("for loop")
		// fmt.Println("emptyMessageCounter", emptyMessageCounter)
		// fmt.Println(a.wavBuffer)
		select {
		case <-sigs:
			cancel()
			break waitFor

		case <-a.ctx.Done():
			cancel()
			break waitFor

		case s := <-record.WavCh:
			log.DEBUG("smarty read from wav chanel")
			txt, err := a.recognizeName.Recognize(s)
			if err != nil {
				log.ERROR(err.Error())
			}
			a.AddWavToBuf(&wavBuffer{buf: &s, text: txt})
			if txt != "" {
				notEmptyMessageCounter += 1
				emptyMessageCounter = 0
			}
			if txt == "" {
				if notEmptyMessageCounter > 0 && !isListenName {
					wavB := a.GetWavFromBuf(notEmptyMessageCounter)
					translateText, err := a.recognizeCommand.Recognize(wavB)
					if err != nil {
						log.ERROR(err.Error())
					}
					log.INFO("Вы сказали: " + translateText)
					a.runCommand(translateText)
				}
				if emptyMessageCounter > CMD_MAX_EMPTY_MESSAGE && !isListenName {
					isListenName = true
					a.PostSignalEvent(AEStartListeningName)
				}
				notEmptyMessageCounter = 0
				emptyMessageCounter += 1
			}
			if isListenName && len(a.wavBuffer) > 1 && notEmptyMessageCounter > 1 {
				wavB := a.GetWavFromBuf(2)
				textWithName, err := a.recognizeName.Recognize(wavB)
				if err != nil {
					log.ERROR(err.Error())
				}
				if IsFindedNameInText(a.Names, textWithName) {
					isListenName = false
					a.PostSignalEvent(AEStartListeningCommand)
				}

			}
		}
	}
}

func (a *Assiser) Print(t ...any) {
	fmt.Println(a.Names[0]+": ", fmt.Sprint(t...))
}

func (a *Assiser) PostSignalEvent(s AssiserEvent) {
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

type TypeCommand string

const (
	tcExec TypeCommand = "exec"
)

type ObjectCommand struct {
	Type     TypeCommand `json:"type"`
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
	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		return err
	}
	cByte, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	var data []ObjectCommand
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

func (a *Assiser) AddGenCommand(data ObjectCommand) {
	var f CommandFunc
	if data.Type == tcExec {
		f = a.newCommandExec(data.Path, data.Args...)
	}
	a.AddCommand(data.Commands, f)
}

func IsFindedNameInText(names []string, text string) bool {
	for _, name := range names {
		if fuzzy.TokenSetRatio(name, text) == 100 {
			return true
		}
	}
	return false
}

func CleanCommandFromName(names []string, command string) string {
	res := command
	for _, name := range names {
		res = strings.ReplaceAll(command, name, "")
	}

	return res
}
