import React, { useState } from "react";
import LayoutPage from "../../components/layoutPage";
import { Box, Button, Input, Typography } from "@mui/material";
import * as api from "../../api/api";

const ComponentInput = ({name, value, title, onChange}) => {
    return <Box>
        <label>{title}</label>
        <Input name={name} onChange={onChange} value={value} sx={{ width: 300, mb: 1, ml: 1, borderColor: "white", color: "white" }}/>
    </Box>
}

function PageCommands({back}) {
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

    return <LayoutPage back={back} title={"Commands"}>
        <Box sx={{display: 'flex', flexDirection: 'column', alignItems:'flex-start'}}>
            <ComponentInput title={"Command"} name={"command"} onChange={onChangeForm} value={form['command']}/>
            <ComponentInput title={"Type"} name={"type"} onChange={onChangeForm} value={form['type']}/>
            <ComponentInput title={"Path"} name={"path"} onChange={onChangeForm} value={form['path']}/>
            <ComponentInput title={"Args"} name={"args"} onChange={onChangeForm} value={form['args']}/>
        </Box>
        <Button onClick={submitForm}>Добавить</Button>
    </LayoutPage>
}

export default PageCommands