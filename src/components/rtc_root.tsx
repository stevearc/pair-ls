import * as React from "react";
import LinearProgress from "@mui/material/LinearProgress";
import Header from "./header";
import Window from "./window";
import Snackbar from "./snackbar";
import RTCSignalFlow from "./rtc_signal_flow";
import RTCStaticFlow from "./rtc_static_flow";
import { AppContext } from "../state";
import RTCClient from "../rtc_client";
const { useEffect, useContext } = React;

export default function Root(_: {}) {
  const { client, setClient, dispatch, state } = useContext(AppContext);

  useEffect(() => {
    if (client == null) {
      const newClient = new RTCClient(dispatch);
      setClient(newClient);
    }
  }, [client]);

  if (client == null) {
    return null;
  }

  return (
    <React.Fragment>
      <Header />
      <ConnectionFlow client={client as RTCClient}>
        <React.Suspense fallback={<LinearProgress />}>
          {state.file_id != null && <Window file_id={state.file_id} />}
        </React.Suspense>
      </ConnectionFlow>
      <Snackbar />
    </React.Fragment>
  );
}

type Props = {
  client: RTCClient;
  children: JSX.Element;
};
function ConnectionFlow({ children, client }: Props) {
  const status = client.useStatus();
  if (status === "connected") {
    return <React.Fragment>{children}</React.Fragment>;
  }
  return process.env.IS_STATIC ? (
    <RTCStaticFlow client={client} />
  ) : (
    <RTCSignalFlow client={client} />
  );
}
