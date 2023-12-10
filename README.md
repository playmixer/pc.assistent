

## Package
```
go get github.com/playmixer/pc.assistent
```


## DEMO

### Создаём Dockerfile для запуска сервера распознования речи
```docker
FROM alphacep/kaldi-vosk-server:latest

ENV MODEL vosk-model-ru-0.22
RUN mkdir /opt/vosk-model \
   && cd /opt/vosk-model \
   && wget -q https://alphacephei.com/vosk/models/${MODEL}.zip \
   && unzip ${MODEL}.zip \
   && mv ${MODEL} model \
   && rm -rf model/extra \
   && rm -rf ${MODEL}.zip

EXPOSE 2700
WORKDIR /opt/vosk-server/websocket
CMD [ "python3", "./asr_server.py", "/opt/vosk-model/model" ]
```

### Запускаем vosk-server
```
docker build -t vosk-server . && docker run -p 2700:2700 -it vosk-server
```

### Создаем голосового помощника
```
go get github.com/playmixer/pc.assistent
```
```golang
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/playmixer/pc.assistent/pkg/logger"
	"github.com/playmixer/pc.assistent/pkg/smarty"
	voskclient "github.com/playmixer/pc.assistent/pkg/vosk-client"
)

func main() {

	ctx := context.TODO()

	log := logger.New("app")
	log.LogLevel = logger.INFO

	recognizer := voskclient.New(log)

	assistent := smarty.New(ctx)
	assistent.SetRecognizeCommand(recognizer)
	assistent.SetRecognizeName(recognizer)
	assistent.SetLogger(log)
	assistent.SetConfig(smarty.Config{
		Names:         []string{"альфа", "бета", "бэта"},
		ListenTimeout: time.Second * 1,
	})

	// Голосовые команды
	assistent.AddCommand([]string{"который час"}, func(ctx context.Context, a *smarty.Assiser) {
		txt := fmt.Sprint("Текущее время:", time.Now().Format(time.TimeOnly))
		a.Print(txt)
	})

	log.INFO("Starting App")
	assistent.Start()

	log.INFO("Stop App")
}

```
### Запускаем
```golang
go run .
```