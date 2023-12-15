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

export const AddCommand = async (payload) => {
    return await _post("/command", payload)
}