/* eslint global-require: off, import/no-extraneous-dependencies: off */

const developmentEnvironments = ["development", "test"];

module.exports = (api) => {
  const development = api.env(developmentEnvironments);

  return {
    presets: [
      // @babel/preset-env will automatically target our browserslist targets
      require("@babel/preset-env"),
      require("@babel/preset-typescript"),
      [require("@babel/preset-react"), { development }],
    ],
    plugins: [["@babel/transform-runtime"]],
  };
};
