import { useState, useEffect } from 'react';

import './App.css';
import websocket from './components/websocket';


const WS_URL = 'ws://127.0.0.1:8080';

function App() {
  const [isConnected, setConnected] = useState(false)
  const [event, setEvent] = useState(20)
  const [state, setState] = useState({})
  
  const connect = websocket(
    WS_URL, {
      onOpen: (e) => {
        setConnected(true)
        if (state.timerId) {
          clearTimeout(state.timerId)
        }
      },
      onClose: (e) => {
        setConnected(false)
        setState(s => ({
          ...s,
          timerId: setTimeout(function() {
            connect()
            console.log("connecting ws")
          }, 3000)
        }))
      },
      onError: (e) => console.log(e.data),
      onMessage: (e) => {
        if (e.data) {
          var body = JSON.parse(e.data)
          if (body) {
            console.log(body)
            setEvent(body.event)
          }
        }
      }
    }
  )


  useEffect(() => {
    connect()
    console.log("connecting ws")

    return () => {
    }
  }, [])

  return (
    <div className={`App ${!isConnected && "disconnected"}`}>
      {!isConnected && <span>Нет соединения</span>}
      <div className={`box-effect ${event == 30 && "cmd-exec"}`}>
        {((event == 10 || event == 30) && isConnected) 
          ? <div className="listen-effect"></div>
          : <div className="white-effect"></div>
        }
      </div>
    </div>
  );
}

export default App;
