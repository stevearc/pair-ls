import * as React from "react";
import Snackbar from "@mui/material/Snackbar";
import MuiAlert, { AlertProps } from "@mui/material/Alert";
import styled from "@emotion/styled";
import { AppContext, AlertWrapper } from "../state";
const { useContext, useEffect, useState } = React;

export default function PairSnackbar() {
  const { state } = useContext(AppContext);
  return <SnackbarComponent alerts={state.alerts} />;
}

const SnackList = styled.div`
  display: flex;
  flex-direction: column;
  overflow: auto;
  gap: 8px;
`;

function SnackbarComponent_({ alerts }: { alerts: AlertWrapper[] }) {
  return (
    <Snackbar
      open={alerts.length > 0}
      anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
    >
      <SnackList>
        {alerts.map((a) => (
          <DynamicAlert key={a.id} elevation={6} variant="filled" {...a.props}>
            {a.text}
          </DynamicAlert>
        ))}
      </SnackList>
    </Snackbar>
  );
}

const SnackbarComponent = React.memo(SnackbarComponent_);

type DynamicAlertProps = Omit<AlertProps, "children"> & {
  children: string | (() => string);
};
function DynamicAlert(props: DynamicAlertProps) {
  const initialText =
    typeof props.children === "function" ? props.children() : props.children;
  const [text, setText] = useState(initialText);
  useEffect(() => {
    let mounted = true;
    if (typeof props.children === "function") {
      setTimeout(() => {
        if (mounted) {
          setText(
            typeof props.children === "function"
              ? props.children()
              : props.children
          );
        }
      }, 1000);
    }
    return () => {
      mounted = false;
    };
  }, [text]);
  return <MuiAlert {...props}>{text}</MuiAlert>;
}
