package listen

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"strconv"
	"sync"
	"time"

	pvrecorder "github.com/Picovoice/pvrecorder/binding/go"
	"github.com/go-audio/wav"
	concatwav "github.com/moutend/go-wav"
)

type logger interface {
	ERROR(v ...string)
	INFO(v ...string)
	DEBUG(v ...string)
}

type tLog struct{}

func (l *tLog) ERROR(v ...string) {
	fmt.Println(v)
}
func (l *tLog) INFO(v ...string) {
	fmt.Println(v)
}
func (l *tLog) DEBUG(v ...string) {
	fmt.Println(v)
}

type Listener struct {
	NameApp    string
	WavCh      chan []byte
	Long       time.Duration
	stopCh     chan struct{}
	SampleRate int
	BitDepth   int
	NumChans   int
	Filename   string
	log        logger
	IsActive   bool
	StartTime  time.Time
	sliceCh    chan int
	service    sync.Mutex
	sync.Mutex
}

func New(t time.Duration) *Listener {
	return &Listener{
		Long:       t,
		SampleRate: 16000,
		BitDepth:   16,
		NumChans:   1,
		Filename:   "",
		stopCh:     make(chan struct{}),
		WavCh:      make(chan []byte, 1),
		sliceCh:    make(chan int, 1),
		log:        &tLog{},
	}
}

func (l *Listener) SetLogger(log logger) {
	l.log = log
}

func (l *Listener) SetName(name string) {
	l.NameApp = name
}

func (l *Listener) Stop() {
	if !l.IsActive {
		return
	}
	l.log.DEBUG(l.NameApp + ": Stop")
	close(l.stopCh)
	l.Lock()
	defer l.Unlock()
	l.IsActive = false
}

func (l *Listener) SliceRecod() {
	l.sliceCh <- 1
}

func (l *Listener) Start(ctx context.Context) {
	go func() {
		l.service.Lock()
		defer l.service.Unlock()
		if l.IsActive {
			return
		}
		l.StartTime = time.Now()
		l.IsActive = true
		l.stopCh = make(chan struct{})
		// l.WavCh = make(chan []byte, 1)
		showAudioDevices := false
		audioDeviceIndex := -1
		flag.Parse()
		l.log.DEBUG(fmt.Sprintf(l.NameApp+": pvrecorder.go version: %s", pvrecorder.Version))

		if showAudioDevices {
			l.log.DEBUG(l.NameApp + ": Printing devices...")
			if devices, err := pvrecorder.GetAvailableDevices(); err != nil {
				log.Fatalf(l.NameApp+" Error: %s.\n", err.Error())
			} else {
				for i, device := range devices {
					l.log.DEBUG(fmt.Sprintf(l.NameApp+"index: %d, device name: %s", i, device))
				}
			}
			return
		}

		recorder := &pvrecorder.PvRecorder{
			DeviceIndex:         audioDeviceIndex,
			FrameLength:         512,
			BufferedFramesCount: 10,
		}

		l.log.DEBUG(l.NameApp + ": Initializing...")
		if err := recorder.Init(); err != nil {
			l.log.ERROR(l.NameApp+" Error: %s.\n", err.Error())
		}
		defer recorder.Delete()

		l.log.DEBUG(fmt.Sprintf(l.NameApp+": Using device: %s", recorder.GetSelectedDevice()))

		l.log.INFO(l.NameApp + ": Starting listener...")
		if err := recorder.Start(); err != nil {
			l.log.ERROR(l.NameApp+" Error: %s.\n", err.Error())
		}

		l.stopCh = make(chan struct{})
		waitCh := make(chan struct{})

		go func() {
			<-l.stopCh
			l.log.DEBUG(l.NameApp + ": stop chan")
			close(waitCh)
		}()

		var outputWav *wav.Encoder
		outputFile := &WriterSeeker{}
		defer outputFile.Close()
		outputWav = wav.NewEncoder(outputFile, pvrecorder.SampleRate, l.BitDepth, l.NumChans, 1)
		defer outputWav.Close()
		delay := time.NewTicker(l.Long)

	waitLoop:
		for {
			select {
			case <-ctx.Done():
				// l.Lock()
				l.log.DEBUG(l.NameApp + ": Stopping...")
				l.WavCh <- outputFile.buf.Bytes()
				// l.Unlock()
				break waitLoop

			case <-waitCh:
				// l.Lock()
				l.log.DEBUG(l.NameApp + ": Stopping...")
				l.WavCh <- outputFile.buf.Bytes()
				// l.Unlock()
				break waitLoop

			//отрезаем по таймауту
			case <-delay.C:
				// l.Lock()
				l.log.DEBUG(l.NameApp + ": step delay...")
				outputWav.Close()
				outputFile.Close()
				l.log.DEBUG(l.NameApp+": step delay 1 ...", "size buf", strconv.Itoa(outputFile.buf.Len()))
				l.WavCh <- outputFile.buf.Bytes()
				l.log.DEBUG(l.NameApp + ": step delay 1 writed to wav chanel")
				l.log.DEBUG(l.NameApp + ": step delay 2...")
				outputFile = &WriterSeeker{}
				outputWav = wav.NewEncoder(outputFile, pvrecorder.SampleRate, l.BitDepth, l.NumChans, 1)
				l.log.DEBUG(l.NameApp + ": ...stop step delay 2")
				// l.Unlock()

			//отрезаем кусок по команде
			case <-l.sliceCh:
				// l.Lock()
				l.log.DEBUG(l.NameApp+": listener", "slice record")
				l.log.DEBUG(l.NameApp + ": step slice...")
				outputWav.Close()
				outputFile.Close()
				l.log.DEBUG(l.NameApp+": step slice 1...", "size buf", strconv.Itoa(outputFile.buf.Len()))
				l.WavCh <- outputFile.buf.Bytes()
				l.log.DEBUG(l.NameApp + ": step slice 1 writed to wav chanel")
				l.log.DEBUG(l.NameApp + ": step slice 2...")
				outputFile = &WriterSeeker{}
				outputWav = wav.NewEncoder(outputFile, pvrecorder.SampleRate, l.BitDepth, l.NumChans, 1)
				l.log.DEBUG(l.NameApp + ": ...stop step slice 2")
				// l.Unlock()

			default:
				pcm, err := recorder.Read()
				if err != nil {
					l.log.ERROR(fmt.Sprintf(l.NameApp+": Error: %s.\n", err.Error()))
				}
				if outputWav != nil {
					for _, f := range pcm {
						err := outputWav.WriteFrame(f)
						if err != nil {
							l.log.ERROR(err.Error())
						}
					}
				}
			}
		}

		l.log.INFO(l.NameApp + ": Stop listener")
	}()
}

type WriterSeeker struct {
	buf bytes.Buffer
	pos int
}

// Write writes to the buffer of this WriterSeeker instance
func (ws *WriterSeeker) Write(p []byte) (n int, err error) {
	// If the offset is past the end of the buffer, grow the buffer with null bytes.
	if extra := ws.pos - ws.buf.Len(); extra > 0 {
		if _, err := ws.buf.Write(make([]byte, extra)); err != nil {
			return n, err
		}
	}

	// If the offset isn't at the end of the buffer, write as much as we can.
	if ws.pos < ws.buf.Len() {
		n = copy(ws.buf.Bytes()[ws.pos:], p)
		p = p[n:]
	}

	// If there are remaining bytes, append them to the buffer.
	if len(p) > 0 {
		var bn int
		bn, err = ws.buf.Write(p)
		n += bn
	}

	ws.pos += n
	return n, err
}

// Seek seeks in the buffer of this WriterSeeker instance
func (ws *WriterSeeker) Seek(offset int64, whence int) (int64, error) {
	newPos, offs := 0, int(offset)
	switch whence {
	case io.SeekStart:
		newPos = offs
	case io.SeekCurrent:
		newPos = ws.pos + offs
	case io.SeekEnd:
		newPos = ws.buf.Len() + offs
	}
	if newPos < 0 {
		return 0, errors.New("negative result pos")
	}
	ws.pos = newPos
	return int64(newPos), nil
}

// Reader returns an io.Reader. Use it, for example, with io.Copy, to copy the content of the WriterSeeker buffer to an io.Writer
func (ws *WriterSeeker) Reader() io.Reader {
	return bytes.NewReader(ws.buf.Bytes())
}

// Close :
func (ws *WriterSeeker) Close() error {
	return nil
}

// BytesReader returns a *bytes.Reader. Use it when you need a reader that implements the io.ReadSeeker interface
func (ws *WriterSeeker) BytesReader() *bytes.Reader {
	return bytes.NewReader(ws.buf.Bytes())
}

func IsWavEmpty(data []byte) (bool, error) {
	// Проверяем, что у нас есть достаточно данных для анализа заголовка WAV
	if len(data) < 44 {
		return false, fmt.Errorf("недостаточно данных для анализа заголовка WAV")
	}

	// Проверяем сигнатуру "RIFF" и "WAVE"
	if string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return false, fmt.Errorf("неверная сигнатура WAV")
	}

	// Получаем формат аудио из заголовка WAV
	audioFormat := int(data[20]) | int(data[21])<<8

	// Проверяем, что формат аудио - PCM
	if audioFormat != 1 {
		return false, fmt.Errorf("поддерживаются только файлы PCM")
	}

	// Получаем число каналов из заголовка WAV
	numChannels := int(data[22]) | int(data[23])<<8

	// Получаем битность из заголовка WAV
	bitsPerSample := int(data[34]) | int(data[35])<<8

	// Вычисляем размер блока данных сэмплов
	blockSize := (bitsPerSample / 8) * numChannels

	// Анализируем данные сэмплов
	for i := 44; i < len(data); i += blockSize {
		for j := 0; j < blockSize; j++ {
			if data[i+j] != 0 {
				return false, nil
			}
		}
	}

	return true, nil
}

func IsSilent(data []byte) (bool, error) {
	// Проверяем, что у нас есть достаточно данных для анализа заголовка WAV
	if len(data) < 44 {
		return false, fmt.Errorf("недостаточно данных для анализа заголовка WAV")
	}

	// Анализируем данные сэмплов
	for i := 44; i < len(data); i += 2 {
		// Преобразуем два байта в 16-битное число
		sample := int16(data[i]) | int16(data[i+1])<<8

		// Проверяем, является ли амплитуда нулевой
		// fmt.Println("sample", sample, int16(data[i]), int16(data[i+1])<<8)
		if sample != 0 {
			return false, nil
		}
	}

	return true, nil
}

func IsVoiceless(data []byte, minSample, maxSample int16) (bool, int16, error) {
	// Проверяем, что у нас есть достаточно данных для анализа заголовка WAV
	if len(data) < 44 {
		return false, 0, fmt.Errorf("недостаточно данных для анализа заголовка WAV")
	}

	// Анализируем данные сэмплов
	for i := 44; i < len(data); i += 2 {
		// Преобразуем два байта в 16-битное число
		sample := int16(data[i]) | int16(data[i+1])<<8

		// Проверяем, является ли амплитуда нулевой
		// fmt.Println("sample", sample)
		if sample > minSample && sample < maxSample {
			return true, sample, nil
		}
	}

	return false, 0, nil
}

func ConcatWav(i1, i2 []byte) []byte {

	// Create wav.File.
	a := &concatwav.File{}
	b := &concatwav.File{}

	// Unmarshal input1.wav and input2.wav.
	concatwav.Unmarshal(i1, a)
	concatwav.Unmarshal(i2, b)

	// Add input2.wav to input1.wav.
	c, _ := concatwav.New(a.SamplesPerSec(), a.BitsPerSample(), a.Channels())
	io.Copy(c, a)
	io.Copy(c, b)

	// Marshal input1.wav and save result.
	file, _ := concatwav.Marshal(c)
	return file
}
