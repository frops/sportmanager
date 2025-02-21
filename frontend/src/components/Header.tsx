import SportsSoccerIcon from '@mui/icons-material/SportsSoccer';
import { AppBar, Toolbar, Typography } from '@mui/material';
import React from 'react';

const Header: React.FC = () => {
  return (
    <AppBar position="static">
      <Toolbar>
        <SportsSoccerIcon sx={{ mr: 2 }} />
        <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
          Sport Manager
        </Typography>
      </Toolbar>
    </AppBar>
  );
};

export default Header; 