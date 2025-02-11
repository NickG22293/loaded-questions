import { useEffect, useState } from "react";
import { useParams, useNavigate, useLocation } from "react-router-dom";

function GameLobby() {
  const { sessionID } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const playerID = location.state?.playerID || null;
  const [players, setPlayers] = useState([]);
  const [gameStarted, setGameStarted] = useState(false);


  useEffect(() => {
    const fetchPlayers = async () => {
      const response = await fetch(`http://localhost:8080/session/${sessionID}`);
      if (response.ok) {
        const data = await response.json();
        setPlayers(Object.values(data.players)); // Convert object to array
      }
    };

    fetchPlayers();
    const interval = setInterval(fetchPlayers, 3000); // Poll every 3 sec
    return () => clearInterval(interval);
  }, [sessionID]);

  // Redirect players when the game starts
  useEffect(() => {
    if (gameStarted) {
      navigate(`/session/${sessionID}/game`, { state: { playerID } });
    }
  }, [gameStarted, navigate, sessionID, playerID]);

  const startGame = async () => {
    const response = await fetch(`http://localhost:8080/session/${sessionID}/start`, {
      method: "POST",
    });

    if (response.ok) {
      setGameStarted(true)
      navigate(`/session/${sessionID}/game`);
    } else {
      alert("Failed to start game.");
    }
  };

  return (
    <div>
      <h1>Game Lobby</h1>
      <p>Session ID: {sessionID}</p>
      
      <h2>Players Joined:</h2>
      <ul>
        {players.map((player) => (
          <li key={player.id}>{player.name}</li>
        ))}
      </ul>

      {/* Show "Join as Player" button only if user hasn't joined */}
      {!playerID && (
        <button onClick={() => navigate(`/session/${sessionID}`)}>Join as Player</button>
      )}

      {/* Show "Start Game" button only if at least 3 players have joined */}
      {players.length >= 3 && (
        <button onClick={startGame}>Start Game</button>
      )}
    </div>
  );
}

export default GameLobby;
