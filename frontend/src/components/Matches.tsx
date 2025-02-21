import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import LocationOnIcon from '@mui/icons-material/LocationOn';
import PeopleIcon from '@mui/icons-material/People';
import SportsSoccerIcon from '@mui/icons-material/SportsSoccer';
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Fab,
  Grid,
  IconButton,
  Stack,
  TextField,
  Typography
} from '@mui/material';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import { DateTimePicker } from '@mui/x-date-pickers/DateTimePicker';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { format } from 'date-fns';
import { enGB } from 'date-fns/locale';
import React, { useEffect, useState } from 'react';
import { config } from '../config';

interface Match {
  id: number;
  date: string;
  location: string;
  venueName: string;
  mapLink: string;
  minPlayers: number;
  maxPlayers: number;
  players: Player[];
  active: boolean;
}

interface Player {
  id: number;
  name: string;
}

interface NewMatch {
  date: Date;
  location: string;
  venueName: string;
  mapLink: string;
  minPlayers: number;
  maxPlayers: number;
}

declare global {
  interface Window {
    Telegram?: {
      WebApp: {
        initDataUnsafe: {
          user?: {
            id: number;
            first_name: string;
            last_name?: string;
            username?: string;
          };
        };
      };
    };
  }
}

const PLAYER_NAME_KEY = 'sportPlayerName';

const getNextSunday = () => {
  const today = new Date();
  const day = today.getDay();
  const diff = day === 0 ? 0 : 7 - day; // If today is Sunday, use today
  const nextSunday = new Date(today);
  nextSunday.setDate(today.getDate() + diff);
  nextSunday.setHours(18, 0, 0, 0); // Set to 18:00
  return nextSunday;
};

const Matches: React.FC = () => {
  const [matches, setMatches] = useState<Match[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [matchToDelete, setMatchToDelete] = useState<number | null>(null);
  const [playerNameDialogOpen, setPlayerNameDialogOpen] = useState(false);
  const [playerName, setPlayerName] = useState('');
  const [matchToJoin, setMatchToJoin] = useState<number | null>(null);
  const [newMatch, setNewMatch] = useState<NewMatch>({
    date: getNextSunday(),
    location: '',
    venueName: '',
    mapLink: '',
    minPlayers: 10,
    maxPlayers: 12,
  });
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  useEffect(() => {
    fetchMatches();

    // Try to get name from Telegram WebApp
    const telegramUser = window.Telegram?.WebApp.initDataUnsafe.user;
    if (telegramUser) {
      const name = telegramUser.username ||
        `${telegramUser.first_name}${telegramUser.last_name ? ` ${telegramUser.last_name}` : ''}`;
      setPlayerName(name);
      localStorage.setItem(PLAYER_NAME_KEY, name);
    } else {
      // If not in Telegram, try to get from localStorage
      const savedName = localStorage.getItem(PLAYER_NAME_KEY);
      if (savedName) {
        setPlayerName(savedName);
      }
    }
  }, []);

  useEffect(() => {
    if (createDialogOpen && matches.length > 0) {
      const activeMatches = matches.filter(m => m.active);
      if (activeMatches.length > 0) {
        const lastMatch = activeMatches[activeMatches.length - 1];
        setNewMatch(prev => ({
          ...prev,
          venueName: lastMatch.venueName || 'Nova Sports Soccer Field',
          mapLink: lastMatch.mapLink || ''
        }));
      }
    }
  }, [createDialogOpen, matches]);

  const fetchMatches = async () => {
    try {
      setLoading(true);
      const response = await fetch(`${config.backendUrl}/api/matches`);
      const data = await response.json();
      setMatches(data);
      setError(null);
    } catch (error) {
      console.error('Error fetching matches:', error);
      setError('Failed to load matches. Please try again later.');
    } finally {
      setLoading(false);
    }
  };

  const handleJoinMatch = async (matchId: number) => {
    const savedName = localStorage.getItem(PLAYER_NAME_KEY);
    if (!savedName) {
      setMatchToJoin(matchId);
      setPlayerNameDialogOpen(true);
      return;
    }

    try {
      const telegramUser = window.Telegram?.WebApp.initDataUnsafe.user;
      const response = await fetch(`${config.backendUrl}/api/matches/${matchId}/join`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: savedName,
          telegramId: telegramUser?.id
        }),
      });
      if (!response.ok) {
        throw new Error('Failed to join match');
      }
      await fetchMatches();
      setSuccessMessage('Successfully joined the match!');
    } catch (error) {
      console.error('Error joining match:', error);
      setError('Failed to join match. Please try again.');
    }
  };

  const handleLeaveMatch = async (matchId: number) => {
    const savedName = localStorage.getItem(PLAYER_NAME_KEY);
    if (!savedName) return;

    try {
      const response = await fetch(`${config.backendUrl}/api/matches/${matchId}/leave`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name: savedName }),
      });
      if (!response.ok) {
        throw new Error('Failed to leave match');
      }
      await fetchMatches();
      setSuccessMessage('Successfully left the match!');
    } catch (error) {
      console.error('Error leaving match:', error);
      setError('Failed to leave match. Please try again.');
    }
  };

  const handleRestoreMatch = async (matchId: number) => {
    try {
      const response = await fetch(`${config.backendUrl}/api/matches/${matchId}/restore`, {
        method: 'POST',
      });
      if (!response.ok) {
        throw new Error('Failed to restore match');
      }
      await fetchMatches();
      setSuccessMessage('Match successfully restored');
    } catch (error) {
      console.error('Error restoring match:', error);
      setError('Failed to restore match. Please try again.');
    }
  };

  const handleDeleteMatch = async () => {
    if (matchToDelete === null) return;

    try {
      const response = await fetch(`${config.backendUrl}/api/matches/${matchToDelete}`, {
        method: 'DELETE',
      });
      if (!response.ok) {
        throw new Error('Failed to delete match');
      }
      await fetchMatches();
      setSuccessMessage('Match successfully deleted');
    } catch (error) {
      console.error('Error deleting match:', error);
      setError('Failed to delete match. Please try again.');
    } finally {
      setDeleteDialogOpen(false);
      setMatchToDelete(null);
    }
  };

  const handlePlayerNameSubmit = async () => {
    if (!playerName.trim()) {
      setError('Please enter your name');
      return;
    }

    localStorage.setItem(PLAYER_NAME_KEY, playerName);
    setPlayerNameDialogOpen(false);

    if (matchToJoin !== null) {
      await handleJoinMatch(matchToJoin);
      setMatchToJoin(null);
    }
  };

  const handleCreateMatch = async () => {
    try {
      const response = await fetch(`${config.backendUrl}/api/matches`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(newMatch),
      });
      if (!response.ok) {
        throw new Error('Failed to create match');
      }
      await fetchMatches();
      setCreateDialogOpen(false);
      setSuccessMessage('Match created successfully!');
      setNewMatch({
        date: getNextSunday(),
        location: '',
        venueName: '',
        mapLink: '',
        minPlayers: 10,
        maxPlayers: 12,
      });
    } catch (error) {
      console.error('Error creating match:', error);
      setError('Failed to create match. Please try again.');
    }
  };

  const getChipColor = (playersCount: number, minPlayers: number, maxPlayers: number) => {
    if (playersCount >= maxPlayers) return 'error';
    if (playersCount >= minPlayers) return 'success';
    return 'warning';
  };

  if (loading) {
    return (
      <Box sx={{ textAlign: 'center', mt: 4 }}>
        <SportsSoccerIcon sx={{ fontSize: 60, color: 'primary.main', animation: 'spin 2s linear infinite' }} />
        <Typography variant="h6" sx={{ mt: 2 }}>
          Loading matches...
        </Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ position: 'relative', minHeight: '80vh' }}>
      <Typography variant="h4" gutterBottom>
        Upcoming Matches
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      <Grid container spacing={3}>
        {matches.length === 0 ? (
          <Grid item xs={12}>
            <Card>
              <CardContent sx={{ textAlign: 'center', py: 4 }}>
                <SportsSoccerIcon sx={{ fontSize: 60, color: 'text.secondary', mb: 2 }} />
                <Typography variant="h6" color="text.secondary" gutterBottom>
                  No matches scheduled
                </Typography>
                <Typography color="text.secondary" sx={{ mb: 2 }}>
                  Be the first to create a match!
                </Typography>
                <Button
                  variant="contained"
                  startIcon={<AddIcon />}
                  onClick={() => setCreateDialogOpen(true)}
                >
                  Create Match
                </Button>
              </CardContent>
            </Card>
          </Grid>
        ) : (
          matches.map((match) => (
            <Grid
              item
              xs={12}
              md={match.active ? 6 : 12}
              key={match.id}
              sx={match.active ? {} : { opacity: 0.6 }}
            >
              <Card sx={match.active ? {} : { backgroundColor: 'grey.100' }}>
                <CardContent sx={match.active ? {} : { py: 1 }}>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                    <Typography variant={match.active ? 'h6' : 'body1'} gutterBottom>
                      {format(new Date(match.date), 'EEEE, d MMMM yyyy, HH:mm', { locale: enGB })}
                    </Typography>
                    {match.active ? (
                      <IconButton
                        color="error"
                        onClick={() => {
                          setMatchToDelete(match.id);
                          setDeleteDialogOpen(true);
                        }}
                      >
                        <DeleteIcon />
                      </IconButton>
                    ) : (
                      <Button
                        size="small"
                        onClick={() => handleRestoreMatch(match.id)}
                      >
                        Restore
                      </Button>
                    )}
                  </Box>
                  {match.active ? (
                    <>
                      <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                        <LocationOnIcon sx={{ mr: 1, color: 'text.secondary' }} />
                        <Typography color="text.secondary">
                          {match.venueName}
                        </Typography>
                      </Box>
                      <Stack spacing={2}>
                        <Box>
                          <Chip
                            icon={<PeopleIcon />}
                            label={`${match.players.length}/${match.maxPlayers} players`}
                            color={getChipColor(match.players.length, match.minPlayers, match.maxPlayers)}
                          />
                          <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                            {match.players.length > 0
                              ? match.players.map(p => p.name).join(', ')
                              : 'No players yet'}
                          </Typography>
                        </Box>
                        <Box sx={{ display: 'flex', gap: 1 }}>
                          {match.players.some(p => p.name === localStorage.getItem(PLAYER_NAME_KEY)) ? (
                            <Button
                              variant="outlined"
                              color="error"
                              onClick={() => handleLeaveMatch(match.id)}
                            >
                              Leave Match
                            </Button>
                          ) : (
                            <Button
                              variant="contained"
                              color="primary"
                              onClick={() => handleJoinMatch(match.id)}
                              disabled={match.players.length >= match.maxPlayers}
                            >
                              {match.players.length >= match.maxPlayers ? 'Full' : 'Join Match'}
                            </Button>
                          )}
                          <Button
                            variant="outlined"
                            color="primary"
                            href={match.mapLink}
                            target="_blank"
                            rel="noopener noreferrer"
                            startIcon={<LocationOnIcon />}
                          >
                            Location
                          </Button>
                        </Box>
                      </Stack>
                    </>
                  ) : (
                    <Typography color="error" variant="body2">
                      Cancelled
                    </Typography>
                  )}
                </CardContent>
              </Card>
            </Grid>
          ))
        )}
      </Grid>

      <Fab
        color="primary"
        sx={{ position: 'fixed', bottom: 16, right: 16 }}
        onClick={() => setCreateDialogOpen(true)}
      >
        <AddIcon />
      </Fab>

      <Dialog
        open={createDialogOpen}
        onClose={() => setCreateDialogOpen(false)}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>Create New Match</DialogTitle>
        <DialogContent>
          <Box sx={{ mt: 2 }}>
            <LocalizationProvider dateAdapter={AdapterDateFns} adapterLocale={enGB}>
              <DateTimePicker
                label="Date and Time"
                value={newMatch.date}
                onChange={(newValue) => setNewMatch({ ...newMatch, date: newValue || getNextSunday() })}
                sx={{ width: '100%', mb: 2 }}
                format="EEEE, d MMMM yyyy, HH:mm"
                ampm={false}
              />
            </LocalizationProvider>
            <TextField
              fullWidth
              label="Venue Name"
              value={newMatch.venueName}
              onChange={(e) => setNewMatch({ ...newMatch, venueName: e.target.value })}
              sx={{ mb: 2 }}
            />
            <TextField
              fullWidth
              label="Location Link (Google Maps)"
              value={newMatch.mapLink}
              onChange={(e) => setNewMatch({ ...newMatch, mapLink: e.target.value })}
              sx={{ mb: 2 }}
              placeholder="https://maps.google.com/?q=..."
            />
            <TextField
              fullWidth
              label="Minimum Players"
              type="number"
              value={newMatch.minPlayers}
              onChange={(e) => setNewMatch({ ...newMatch, minPlayers: parseInt(e.target.value) })}
              sx={{ mb: 2 }}
            />
            <TextField
              fullWidth
              label="Maximum Players"
              type="number"
              value={newMatch.maxPlayers}
              onChange={(e) => setNewMatch({ ...newMatch, maxPlayers: parseInt(e.target.value) })}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateDialogOpen(false)}>Cancel</Button>
          <Button onClick={handleCreateMatch} variant="contained">
            Create
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={playerNameDialogOpen}
        onClose={() => setPlayerNameDialogOpen(false)}
      >
        <DialogTitle>Enter Your Name</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Your Name"
            fullWidth
            value={playerName}
            onChange={(e) => setPlayerName(e.target.value)}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setPlayerNameDialogOpen(false)}>Cancel</Button>
          <Button onClick={handlePlayerNameSubmit} variant="contained">
            Save
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={deleteDialogOpen}
        onClose={() => setDeleteDialogOpen(false)}
      >
        <DialogTitle>Delete Match</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Are you sure you want to delete this match? This action cannot be undone.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialogOpen(false)}>Cancel</Button>
          <Button onClick={handleDeleteMatch} color="error" variant="contained">
            Delete
          </Button>
        </DialogActions>
      </Dialog>

      <Alert
        severity="success"
        sx={{
          position: 'fixed',
          bottom: 16,
          left: '50%',
          transform: 'translateX(-50%)',
          display: successMessage ? 'flex' : 'none',
        }}
        onClose={() => setSuccessMessage(null)}
      >
        {successMessage}
      </Alert>
    </Box>
  );
};

export default Matches; 