import { useEffect, useState } from "react";
import { useParams, useNavigate, useLocation } from "react-router-dom";

function GamePage() {
  const { sessionID } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const playerID = location.state?.playerID || null;
  const [gameState, setGameState] = useState(null);
  const [answer, setAnswer] = useState("");
  const [isAsker, setIsAsker] = useState(false);

  useEffect(() => {
    const fetchGameState = async () => {
      const response = await fetch(`http://localhost:8080/session/${sessionID}`);
      if (response.ok) {
        const data = await response.json();
        setGameState(data);
        setIsAsker(data.asker?.id === playerID);
      }
    };

    fetchGameState();
    const interval = setInterval(fetchGameState, 3000);
    return () => clearInterval(interval);
  }, [sessionID, playerID]);

  const submitAnswer = async () => {
    const response = await fetch(`http://localhost:8080/session/${sessionID}/answer`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ player_id: playerID, answer }),
    });

    if (response.ok) {
      setAnswer(""); // Clear the input field after submitting
    } else {
      alert("Failed to submit answer.");
    }
  };

  return (
    <div className="page-container">
      <h1>Loaded Questions - Game Round</h1>

      {gameState ? (
        <>
          <p><strong>Current Question:</strong> {gameState.question || "Waiting for the Asker..."}</p>

          {!isAsker && gameState.question && !gameState.answers[playerID] && (
            <div>
              <input
                type="text"
                placeholder="Enter your answer..."
                value={answer}
                onChange={(e) => setAnswer(e.target.value)}
                className="input"
              />
              <button className="btn-primary" onClick={submitAnswer}>Submit Answer</button>
            </div>
          )}

          {isAsker && !gameState.question && (
            <button className="btn-primary" onClick={() => navigate(`/session/${sessionID}/ask`)}>Ask a Question</button>
          )}
        </>
      ) : (
        <p>Loading game state...</p>
      )}
    </div>
  );
}

export default GamePage;
