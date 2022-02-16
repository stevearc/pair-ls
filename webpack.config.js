const webpack = require("webpack");
const path = require("path");

module.exports = (env, argv) => {
  var config = {
    entry: "./src/index.tsx",
    output: {
      filename: "main.js",
      path: path.resolve(
        __dirname,
        env.IS_STATIC ? "static/dist" : "server/dist"
      ),
    },
    module: {
      rules: [
        {
          test: /\.[tj]sx?$/,
          exclude: /node_modules/,
          use: {
            loader: "babel-loader",
            options: {
              presets: ["@babel/preset-env"],
            },
          },
        },
      ],
    },
    resolve: {
      extensions: [".tsx", ".ts", ".js"],
    },
  };
  if (argv.mode === "development") {
    config.devtool = "inline-source-map";
  }
  config.plugins = [
    new webpack.DefinePlugin({
      "process.env.IS_STATIC": env.IS_STATIC || "false",
    }),
  ];

  return config;
};
