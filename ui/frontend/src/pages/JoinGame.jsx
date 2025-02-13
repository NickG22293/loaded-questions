import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";

function JoinGame() {
  const { sessionID } = useParams();
  const navigate = useNavigate();
  const [name, setName] = useState("");

  const joinSession = async () => {
    if (!name) return;
    
    const response = await fetch(`http://localhost:8080/session/${sessionID}/join?name=${name}`, {
      method: "POST",
    });

    if (response.ok) {
      const data = await response.json();
      navigate(`/session/${sessionID}/lobby`, { state: { playerID: data.player_id } });
    } else {
      alert("Failed to join session.");
    }
  };

  return (
    <div className="page-container">
      <h1>Join Game</h1>
      <p>Game Code: {sessionID}</p>
      <input
        type="text"
        placeholder="Enter your name"
        value={name}
        onChange={(e) => setName(e.target.value)}
        className="input"
      />
      <button className="btn-primary" onClick={joinSession}>Join</button>
    </div>
  );
}

export default JoinGame;
