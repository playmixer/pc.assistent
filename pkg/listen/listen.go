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
	"time"

	pvrecorder "github.com/Picovoice/pvrecorder/binding/go"
	"github.com/go-audio/wav"
)

type logger interface {
	ERROR(v ...string)
	INFO(v ...string)
	DEBUG(v ...string)
}

type Listener struct {
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
	}
}

func (l *Listener) SetLogger(log logger) {
	l.log = log
}

func (l *Listener) Stop() {
	if !l.IsActive {
		return
	}
	l.log.DEBUG("Stop")
	close(l.stopCh)
	l.IsActive = false
}

func (l *Listener) SliceRecod() {
	l.sliceCh <- 1
}

func (l *Listener) Start(ctx context.Context) {
	go func() {
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
		l.log.DEBUG(fmt.Sprintf("pvrecorder.go version: %s", pvrecorder.Version))

		if showAudioDevices {
			l.log.DEBUG("Printing devices...")
			if devices, err := pvrecorder.GetAvailableDevices(); err != nil {
				log.Fatalf("Error: %s.\n", err.Error())
			} else {
				for i, device := range devices {
					l.log.DEBUG(fmt.Sprintf("index: %d, device name: %s", i, device))
				}
			}
			return
		}

		recorder := &pvrecorder.PvRecorder{
			DeviceIndex:         audioDeviceIndex,
			FrameLength:         512,
			BufferedFramesCount: 10,
		}

		l.log.DEBUG("Initializing...")
		if err := recorder.Init(); err != nil {
			l.log.ERROR("Error: %s.\n", err.Error())
		}
		defer recorder.Delete()

		l.log.DEBUG(fmt.Sprintf("Using device: %s", recorder.GetSelectedDevice()))

		l.log.INFO("Starting listener...")
		if err := recorder.Start(); err != nil {
			l.log.ERROR("Error: %s.\n", err.Error())
		}

		l.stopCh = make(chan struct{})
		waitCh := make(chan struct{})

		go func() {
			<-l.stopCh
			l.log.DEBUG("stop chan")
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
				l.log.DEBUG("Stopping...")
				l.WavCh <- outputFile.buf.Bytes()
				break waitLoop

			case <-waitCh:
				l.log.DEBUG("Stopping...")
				l.WavCh <- outputFile.buf.Bytes()
				break waitLoop

				//отрезаем по таймауту
			case <-delay.C:
				l.log.DEBUG("step...")
				outputWav.Close()
				outputFile.Close()
				l.log.DEBUG("step 1...", "size buf", strconv.Itoa(outputFile.buf.Len()))
				l.WavCh <- outputFile.buf.Bytes()
				l.log.DEBUG("step 2...")
				outputFile = &WriterSeeker{}
				outputWav = wav.NewEncoder(outputFile, pvrecorder.SampleRate, 16, 1, 1)
				l.log.DEBUG("...stop step")

				//отрезаем кусок по команде
			case <-l.sliceCh:
				l.log.INFO("listener", "slice record")
				l.log.DEBUG("step...")
				outputWav.Close()
				outputFile.Close()
				l.log.DEBUG("step 1...", "size buf", strconv.Itoa(outputFile.buf.Len()))
				l.WavCh <- outputFile.buf.Bytes()
				l.log.DEBUG("step 2...")
				outputFile = &WriterSeeker{}
				outputWav = wav.NewEncoder(outputFile, pvrecorder.SampleRate, 16, 1, 1)
				l.log.DEBUG("...stop step")

			default:
				// l.log.DEBUG("record read")
				pcm, err := recorder.Read()
				if err != nil {
					l.log.ERROR(fmt.Sprintf("Error: %s.\n", err.Error()))
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

		l.log.INFO("Stop listener")
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
