import * as React from "react";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import LoadingButton from "@mui/lab/LoadingButton";
import TextField from "@mui/material/TextField";
import { AppContext, showToast } from "../state";
import RTCClient from "../rtc_client";
const { useEffect, useContext, useMemo, useState } = React;

type Props = {
  client: RTCClient;
};
export default function RTCStaticFlow({ client }: Props) {
  const token = useMemo(
    () => new URLSearchParams(window.location.search).get("t"),
    []
  );
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
    >
      {token == null ? (
        <CallInitiator client={client} />
      ) : (
        <CallResponder client={client} token={token} />
      )}
    </Box>
  );
}

type InitiatorProps = {
  client: RTCClient;
};
function CallInitiator({ client }: InitiatorProps) {
  const { dispatch } = useContext(AppContext);
  const [serverCode, setServerCode] = useState("");
  const [localDescription, setLocalDescription] = useState<string | null>(null);
  const status = client.useStatus();
  if (localDescription == null) {
    return (
      <React.Fragment>
        <div>Click the button to start a connection</div>
        <LoadingButton
          sx={{ marginTop: "8px" }}
          loading={status === "connecting"}
          variant="contained"
          onClick={async () => {
            try {
              const desc = await client.startCall();
              setLocalDescription(desc);
            } catch (e) {
              showToast(dispatch, `Connection error ${e}`, {
                severity: "error",
              });
            }
          }}
        >
          Connect
        </LoadingButton>
      </React.Fragment>
    );
  }
  const hasError = status === "failed" || status === "no_session";
  const helperText = hasError ? "Failed to establish connection" : null;
  return (
    <React.Fragment>
      <div>Copy this token and send it to the editor</div>
      <TokenBox token={localDescription} />
      <div>The editor will respond with a token. Paste it in here</div>
      <TextField
        label="Editor response"
        sx={{ margin: "8px 0" }}
        helperText={helperText}
        fullWidth
        multiline
        rows={4}
        error={hasError}
        value={serverCode}
        onChange={(e) => {
          const newVal = e.target.value.trim();
          setServerCode(newVal);
          if (newVal !== "") {
            client.setAnswer(e.target.value).catch((e) => {
              showToast(dispatch, `Invalid code: ${e}`, { severity: "error" });
              setServerCode("");
            });
          }
        }}
      />
      <div>
        Not sure what's going on? Read{" "}
        <a target="_blank" href="https://github.com/stevearc/pair-ls#pair-ls">
          the documentation
        </a>
      </div>
    </React.Fragment>
  );
}

type ResponderProps = {
  client: RTCClient;
  token: string;
};
function CallResponder({ client, token }: ResponderProps) {
  const { dispatch } = useContext(AppContext);
  const [localDescription, setLocalDescription] = useState<string | null>(null);
  useEffect(() => {
    client.respondToCall(token).then(setLocalDescription, (e) => {
      showToast(dispatch, `Error creating connection ${e}`, {
        severity: "error",
      });
    });
  }, []);
  if (localDescription == null) {
    return <CircularProgress />;
  }
  return (
    <React.Fragment>
      <div>Paste this token into the editor to connect</div>
      <TokenBox token={localDescription} />
    </React.Fragment>
  );
}

function TokenBox({ token }: { token: string }) {
  const { dispatch } = useContext(AppContext);
  return (
    <React.Fragment>
      <TextField
        label="Connect token"
        sx={{ margin: "8px 0" }}
        fullWidth
        multiline
        rows={4}
        value={token}
        onChange={() => {}}
        onFocus={(e) => {
          e.target.select();
        }}
      />

      {window.isSecureContext && (
        <Button
          variant="outlined"
          sx={{ marginBottom: "8px" }}
          onClick={() => {
            navigator.clipboard.writeText(token).then(
              () => {
                showToast(dispatch, "Copied", {});
              },
              () => {
                showToast(dispatch, "Failed to copy", { severity: "error" });
              }
            );
          }}
        >
          Copy
        </Button>
      )}
    </React.Fragment>
  );
}
