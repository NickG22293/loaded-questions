import { useNavigate } from "react-router-dom";

function App() {
  const navigate = useNavigate();

  const startSession = async () => {
    try { 
      const response = await fetch("http://localhost:8080/session", {
        method: "POST",
      });
      if (response.ok) {
        const data = await response.json();
        navigate(`/session/${data.session_id}/lobby`);
      } else {
        alert("Failed to start session.");
      }
    } catch (e) {
      print(e)
    }
    
  };

  return (
    <div>
      <h1>Welcome to Loaded Questions</h1>
      <button onClick={startSession}>Start Session</button>
    </div>
  );
}

export default App;

