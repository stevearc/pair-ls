import * as React from "react";
import LinearProgress from "@mui/material/LinearProgress";
import Header from "./header";
import Window from "./window";
import Login from "./login";
import Snackbar from "./snackbar";
import { AppContext } from "../state";
import Client from "../client";
const { useCallback, useContext } = React;

export default function Root(_: {}) {
  const { client, dispatch, state, setClient } = useContext(AppContext);
  const handleLogin = useCallback(
    (token) => {
      const proto = window.location.protocol === "https:" ? "wss" : "ws";
      const c = new Client(
        `${proto}://${window.location.host}/client_ws`,
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
