const API = "http://localhost:8080/api/v0"

const _post = async (url, body) => {
    const response = await fetch(API+url, {
        method: "POST",
        headers: {
            "Content-Type": "application/json;charset=utf-8"
        },
        body: JSON.stringify(body)
    })
    return await response.json()
}

const _get = async (url, params) => {
    const _url = new URL(API+url)
    const _keys = Object.keys(params)
    _keys.map((k) => {
        _url.searchParams.append(k, _url[k])
    })
    url = _url.toString()
    const response = await fetch(url, {
        method: "GET",
        headers: {
            "Content-Type": "application/json;charset=utf-8"
        }
    })
    return await response.json()
}

export const AddCommand = async (payload) => {
    return await _post("/command", payload)
}

export const GetCommands = async (payload) => {
    return await _get("/command", {})
}