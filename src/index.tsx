import * as ReactDOM from "react-dom";
import * as React from "react";
import LinearProgress from "@mui/material/LinearProgress";
import CircularProgress from "@mui/material/CircularProgress";
import CssBaseline from "@mui/material/CssBaseline";
import Header from "./components/header";
import Window from "./components/window";
import Login from "./components/login";
import Snackbar from "./components/snackbar";
import { ThemeProvider } from "@mui/material/styles";
import { reducer, getInitialState, AppContext } from "./state";
import { useTheme } from "./colors/colorschemes";
import Client from "./client";
const { useCallback, useContext, useMemo, useReducer, useState } = React;

function Content(_: {}) {
  const { client, dispatch, state, setClient } = useContext(AppContext);
  const handleLogin = useCallback(
    (token) => {
      const proto = window.location.protocol === "https:" ? "wss" : "ws";
      const c = new Client(
        `${proto}://${window.location.host}/ws`,
        dispatch,
        token
      );
      setClient(c);
    },
    [dispatch, setClient]
  );

  return (
    <React.Fragment>
      <Header />
      <React.Suspense fallback={<LinearProgress />}>
        {state.filename != null && <Window filename={state.filename} />}
      </React.Suspense>
      {client == null && <Login onLogin={handleLogin} />}
      <Snackbar />
    </React.Fragment>
  );
}

function App(_: {}) {
  const [state, dispatch] = useReducer(reducer, undefined, getInitialState);
  const [client, setClient] = useState<Client | null>(null);
  const theme = useTheme(state.colorscheme);
  const context = useMemo(
    () => ({
      state,
      dispatch,
      client,
      setClient,
    }),
    [state, dispatch, client, setClient]
  );
  return (
    <React.Fragment>
      <CssBaseline enableColorScheme />
      <ThemeProvider theme={theme}>
        <AppContext.Provider value={context}>
          <React.Suspense fallback={<CircularProgress />}>
            <Content />
          </React.Suspense>
        </AppContext.Provider>
      </ThemeProvider>
    </React.Fragment>
  );
}

(function () {
  const domContainer = document.querySelector("#root");
  ReactDOM.render(<App />, domContainer);
})();
