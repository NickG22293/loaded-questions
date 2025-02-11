import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import App from "./App";
import JoinGame from "./pages/JoinGame.jsx";
import GameLobby from "./pages/GameLobby.jsx";
import GamePage from "./pages/Game.jsx";

ReactDOM.createRoot(document.getElementById("root")).render(
  <React.StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<App />} />
        <Route path="/session/:sessionID" element={<JoinGame />} />
        <Route path="/session/:sessionID/lobby" element={<GameLobby />} />
        <Route path="/session/:sessionID/game" element={<GamePage />} />
      </Routes>
    </BrowserRouter>
  </React.StrictMode>
);