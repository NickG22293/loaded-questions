import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";

function GameLobby() {
  const { sessionID } = useParams();
  const [players, setPlayers] = useState([]);
  const navigate = useNavigate();

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
      <button onClick={() => navigate(`/session/${sessionID}`)}>Join as Player</button>
    </div>
  );
}

export default GameLobby;
