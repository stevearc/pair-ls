import * as React from "react";
import Box from "@mui/material/Box";
import LoadingButton from "@mui/lab/LoadingButton";
import Dialog from "@mui/material/Dialog";
import TextField from "@mui/material/TextField";
import { post } from "../api";
const { useEffect, useState } = React;

type Props = {
  onLogin: (token: string) => void;
};
export default function Login({ onLogin }: Props) {
  const [pass, setPass] = useState("");
  const [connecting, setConnecting] = useState(false);
  const [hasError, setHasError] = useState(false);
  useEffect(() => {
    const storedToken = localStorage.getItem("token") ?? "";
    post<{ token: string }>("login", { password: storedToken }).then((resp) => {
      onLogin(resp.token);
    });
  }, []);
  const submit = async () => {
    setConnecting(true);
    try {
      const resp = await post<{ token: string }>("login", { password: pass });
      localStorage.setItem("token", resp.token);
      onLogin(resp.token);
    } catch {
      setHasError(true);
    } finally {
      setConnecting(false);
    }
  };
  return (
    <Dialog open={true}>
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
          submit();
        }}
      >
        <TextField
          autoFocus
          error={hasError}
          label="Password"
          type="password"
          autoComplete="current-password"
          value={pass}
          onChange={(e) => setPass(e.target.value)}
        />
        <LoadingButton
          sx={{ marginTop: "8px" }}
          loading={connecting}
          variant="contained"
          onClick={submit}
        >
          Login
        </LoadingButton>
      </Box>
    </Dialog>
  );
}
