export const WS_URL = 'ws://127.0.0.1:8080/ws';

export default (url, {onOpen, onClose, onError, onMessage}) => {
    return () => {
        const socket = new WebSocket(url)

        socket.onopen = onOpen
        socket.onclose = onClose
        socket.onmessage = onMessage
        socket.onerror = onError

        return socket
    }
}
