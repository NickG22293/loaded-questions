import PropTypes from 'prop-types';

const PlayerList = ({ sessionID, players }) => {
  return (
    <div className="player-list">
      <h2>Session ID: {sessionID}</h2>
      <ul>
        {Object.values(players).map(player => (
          <li key={player.id}>{player.name}</li>
        ))}
      </ul>
    </div>
  );
};

PlayerList.propTypes = {
  sessionID: PropTypes.string.isRequired,
  players: PropTypes.objectOf(
    PropTypes.shape({
      id: PropTypes.string.isRequired,
      name: PropTypes.string.isRequired,
    })
  ).isRequired,
};

export default PlayerList;
