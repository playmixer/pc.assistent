import { useState, useEffect } from 'react';

import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import MenuIcon from '@mui/icons-material/Menu';
import Button from '@mui/material/Button';

import SensorsOffRoundedIcon from '@mui/icons-material/SensorsOffRounded';
import SettingsRoundedIcon from '@mui/icons-material/SettingsRounded';
import DvrRoundedIcon from '@mui/icons-material/DvrRounded';

import './App.css';
import websocket, { WS_URL } from './components/websocket';

import { red } from '@mui/material/colors';
import PageSettings from './pages/settings';
import PageCommands from './pages/commands';



const pages = [
  {
    title: <DvrRoundedIcon/>,
    name: "commands",
  },
  {
    title: <SettingsRoundedIcon/>,
    name: "settings",
  }
]

function App() {
  const [socket, setSocket] = useState({})
  const [isConnected, setConnected] = useState(false)
  const [anchorEl, setAnchorEl] = useState(null);
  const [event, setEvent] = useState(20)
  const [state, setState] = useState({
    isOpenMenu: false,
    page: "",
    timerId: null,
  })

  
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
        // setState(s => ({
        //   ...s,
        //   timerId: setTimeout(function() {
        //     setSocket(connect())
        //     console.log("connecting ws")
        //   }, 3000)
        // }))
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

  const openPage = (name) => {
    setState(s => ({...s, page: name}))
  }

  const toBack = () => {
    setState(s => ({...s, page: ""}))
  }

  const sendMessage = () => {
    console.log(socket)
    socket?.send("ewqe12345")

  }

  const handleConnect = () => {
    console.log("connecting ws")
    setEvent(20)
    setSocket(connect())
  }

  useEffect(() => {

    return () => {
      setSocket(connect())
      console.log("connecting ws")
    }
  }, [])

  return (
    <div className={`App ${!isConnected && "disconnected"}`}>
      <AppBar position="static">
        <Toolbar>
          <IconButton edge="start" color="inherit" aria-label="menu" sx={{ mr: 2 }} onClick={sendMessage}>
            <MenuIcon />
          </IconButton>
          {!isConnected && <Typography variant="h6" component="div" sx={{ flexGrow: 1, color: red[900] }}><SensorsOffRoundedIcon onClick={handleConnect}/></Typography>}
          <Box sx={{ display: 'flex', justifyContent: 'flex-end', width: "100%" }}>
            <Typography variant="h6" color="inherit" component="div">
              {pages.map((page) => <Button key={page.name} sx={{ color: '#fff' }} onClick={() => openPage(page.name)}>
                {page.title}
              </Button>)}
            </Typography>
          </Box>
        </Toolbar>
      </AppBar>
      {state.page == "" && <div className={`box-effect ${event == 30 && "cmd-exec"}`}>
        {((event == 10 || event == 30) && isConnected) 
          ? <div className="listen-effect"></div>
          : <div className="white-effect"></div>
        }
      </div>}
      {state.page == "settings" && <PageSettings back={toBack}/>}
      {state.page == "commands" && <PageCommands back={toBack}/>}
      
    </div>
  );
}

export default App;
