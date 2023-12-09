
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
