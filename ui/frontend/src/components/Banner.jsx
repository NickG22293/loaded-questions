import PropTypes from 'prop-types';

const Banner = ({ userName }) => {
  return (
    <div className="bg-blue-500 text-white p-4 flex justify-between items-center w-full fixed top-0 left-0">
      <span className="ml-4">Loaded Questions!</span>
      <span className="mr-4">{userName ? userName : 'Guest'}</span>
    </div>
  );
};

Banner.propTypes = {
  userName: PropTypes.string.isRequired,
};

export default Banner;