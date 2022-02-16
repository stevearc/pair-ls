import * as React from "react";
import Box from "@mui/material/Box";
import LoadingButton from "@mui/lab/LoadingButton";
import TextField from "@mui/material/TextField";
import RTCClient from "../rtc_client";
const { useEffect, useState } = React;

type Props = {
  client: RTCClient;
};
export default function RTCSignalFlow({ client }: Props) {
  const status = client.useStatus();
  const [token, setToken] = useState(window.location.pathname.slice(1));
  useEffect(() => {
    if (token !== "") {
      client.connectWithSignalToken(token);
    }
  }, []);
  if (status === "closed") {
    return <div>Connection closed. Please refresh</div>;
  }
  let helperText = null;
  if (status === "no_session") {
    helperText = "Could not find an editing session with that token";
  } else if (status === "failed") {
    helperText = "Failed to establish connection";
  }
  const connect = () => {
    window.history.pushState({ token }, token, "/" + token);
    setToken(token);
    client.connectWithSignalToken(token);
  };

  return (
    <Box
      sx={{
        display: "flex",
        padding: "8px",
        flexDirection: "column",
        alignItems: "center",
        marginTop: "16px",
      }}
      component="form"
      noValidate
      onSubmit={(e: React.SyntheticEvent<HTMLFormElement>) => {
        e.preventDefault();
        connect();
      }}
    >
      <div>Please enter a token to connect to an editing session</div>
      <TextField
        label="Token"
        sx={{ marginTop: "8px" }}
        helperText={helperText}
        error={status === "no_session" || status === "failed"}
        disabled={status === "connecting"}
        value={token}
        onChange={(e) => setToken(e.target.value)}
      />
      <LoadingButton
        sx={{ marginTop: "8px" }}
        disabled={token === ""}
        loading={status === "connecting"}
        variant="contained"
        onClick={() => connect()}
      >
        Connect
      </LoadingButton>
    </Box>
  );
}
