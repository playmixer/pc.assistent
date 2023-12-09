FROM alphacep/kaldi-vosk-server:latest

# ENV RUVERSION 0.22
ENV MODEL vosk-model-ru-0.22
# ENV MODEL vosk-model-small-ru-0.22
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