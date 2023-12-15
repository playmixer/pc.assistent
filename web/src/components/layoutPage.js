import React from "react";

import ArrowBackRoundedIcon from '@mui/icons-material/ArrowBackRounded';
import { Typography } from '@mui/material';
import { Box, Button } from "@mui/material";

function LayoutPage({children, back, title}) {


    return <div>
        <Box sx={{display: 'flex', justifyContent: 'flex-start', mt: 1, mb: 1}}>
            {back && <Button onClick={back} sx={{color: "white"}}><ArrowBackRoundedIcon /></Button>}
            {title && <Box sx={{display: "flex", alignItems: "center", justifyContent: "center",}}>
                <Typography ml={2} color={"white"}>{title}</Typography>
            </Box>}
        </Box>
        <Box sx={{ml: 2, mr: 2}}>
            {children}
        </Box>
    </div>
}

export default LayoutPage