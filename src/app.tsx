import * as React from "react";
import CircularProgress from "@mui/material/CircularProgress";
import CssBaseline from "@mui/material/CssBaseline";
import { ThemeProvider } from "@mui/material/styles";
import { reducer, getInitialState, AppContext } from "./state";
import { useTheme } from "./colors/colorschemes";
import BaseClient from "./base_client";
const { useMemo, useReducer, useState } = React;

export default function App({ children }: { children: JSX.Element }) {
  const [state, dispatch] = useReducer(reducer, undefined, getInitialState);
  const [client, setClient] = useState<BaseClient | null>(null);
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
            {children}
          </React.Suspense>
        </AppContext.Provider>
      </ThemeProvider>
    </React.Fragment>
  );
}
