import React, { useEffect, useState } from "react";
import LayoutPage from "../../components/layoutPage";
import { Box, Button, Input, TextField, Typography } from "@mui/material";
import * as api from "../../api/api";

const ComponentInput = ({name, value, title, onChange}) => {
    return <Box>
        <label>{title}</label>
        <Input name={name} onChange={onChange} value={value} sx={{ width: 300, mb: 1, ml: 1, borderColor: "white", color: "white" }}/>
    </Box>
}

function PageCommands({back}) {
    const [commands, setCommands] = useState([])
    const [form, setForm] = useState({
        type: "",
        command: "",
        path: "",
        args: ""
    })

    const styleInput = {
        width: 300,
        mb: 1,
        ml: 1,
    }

    const onChangeForm = (e) => {
        setForm(s => ({
            ...s,
            [e.target.name]: e.target.value
        }))
    }

    const submitForm = async () => {
        const res = await api.AddCommand(form)
        console.log(res)
    }

    useEffect(() => {
        api.GetCommands()
        .then(r => {
            setCommands(r)
        })

    }, [])

    return <LayoutPage back={back} title={"Commands"}>
        {
            commands.map((command, i) => {
                return <Box key={i} sx={{display: 'flex', flexDirection: 'column', alignItems:'flex-start', mb: 4}}>
                    <ComponentInput title={"Command"} name={"command"} onChange={onChangeForm} value={command.commands?.join("; ")}/>
                    <ComponentInput title={"Type"} name={"type"} onChange={onChangeForm} value={command['type']}/>
                    <ComponentInput title={"Path"} name={"path"} onChange={onChangeForm} value={command['path']}/>
                    <ComponentInput title={"Args"} name={"args"} onChange={onChangeForm} value={command['args']}/>
                </Box>
            })
        }

        <Button onClick={submitForm}>Добавить</Button>
    </LayoutPage>
}

export default PageCommands