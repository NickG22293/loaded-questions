import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import Banner from "../components/Banner";
import PlayerList from "../components/PlayerList";

function GamePage() {
  const { sessionID } = useParams();
  const [gameState, setGameState] = useState(null);
  const [answer, setAnswer] = useState("");
  const [question, setQuestion] = useState("");
  const [isAsker, setIsAsker] = useState(false);
  const playerID = localStorage.getItem("playerID");

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

  const submitQuestion = async () => {
    const response = await fetch(`http://localhost:8080/session/${sessionID}/question`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ player_id: playerID, question }),
    });

    if (response.ok) {
      setQuestion(""); // Clear the input field after submitting
    } else {
      alert("Failed to submit question.");
    }
  };

  return (
    <div className="page-container flex">
      <div className="player-list-container flex flex-col justify-center items-center w-1/4">
        {gameState && <PlayerList sessionID={sessionID} players={gameState.players} />}
      </div>
      <div className="game-ui-container flex flex-col justify-center items-center w-3/4">
        <Banner userName={gameState?.players[playerID]?.name} />
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
              <div>
                <input
                  type="text"
                  placeholder="Enter your question..."
                  value={question}
                  onChange={(e) => setQuestion(e.target.value)}
                  className="input"
                />
                <button className="btn-primary" onClick={submitQuestion}>Submit Question</button>
              </div>
            )}
          </>
        ) : (
          <p>Loading game state...</p>
        )}
      </div>
    </div>
  );
}

export default GamePage;
