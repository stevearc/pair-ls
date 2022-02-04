import * as React from "react";
import { styled } from "@mui/system";
import TabsUnstyled from "@mui/base/TabsUnstyled";
import TabsListUnstyled from "@mui/base/TabsListUnstyled";
import Box from "@mui/material/Box";
import { AppContext, Dispatcher } from "../state";
import TabUnstyled, { tabUnstyledClasses } from "@mui/base/TabUnstyled";
const { useContext, useEffect, useMemo, useRef } = React;

const Tab = styled(TabUnstyled)`
  color: var(--comment);
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: bold;
  background-color: var(--bg);
  margin-top: 8px;
  flex-grow: 1;
  padding: 8px;
  border: none;
  border-radius: 5px 5px 0 0;
  display: flex;
  justify-content: center;

  &.${tabUnstyledClasses.selected} {
    background-color: var(--tab_selected_bg);
    color: var(--tab_selected_fg);
  }

  &:hover {
    background-color: var(--tab_hover_bg);
    color: var(--tab_hover_fg);
  }

  &:active {
    color: var(--tab_hover_fg);
    background-color: var(--tab_hover_bg);
    filter: brightness(0.9);
  }
`;

const TabsList = styled(TabsListUnstyled)`
  background-color: var(--bg);
  margin: 0;
  display: flex;
  align-items: stretch;
  justify-content: flex-start;
  overflow-x: auto;
  height: 100%;
`;

const Tabs = styled(TabsUnstyled)`
  height: 100%;
`;

export default function PublicTabPanel(_: {}) {
  const { state, dispatch } = useContext(AppContext);
  const filenames = useMemo(() => Object.keys(state.files), [state.files]);
  return (
    <TabPanel
      dispatch={dispatch}
      filename={state.filename}
      filenames={filenames}
      viewFile={state.view?.filename}
    />
  );
}

function TabPanel_({
  dispatch,
  filenames,
  filename,
  viewFile,
}: {
  dispatch: Dispatcher;
  filenames: string[];
  filename?: string;
  viewFile?: string;
}) {
  const selectedEl = useRef<null | HTMLButtonElement>(null);
  const shortnames = useShortFilenames(filenames);
  useEffect(() => {
    if (selectedEl.current != null) {
      selectedEl.current.scrollIntoView();
    }
  }, [filename]);
  if (filenames.length === 0) {
    return null;
  }
  return (
    <Box
      sx={{
        minWidth: 0,
        height: "100%",
      }}
    >
      <Tabs
        value={filename}
        onChange={(_event, value) => {
          dispatch({ type: "selectFile", filename: value.toString() });
        }}
      >
        <TabsList>
          {filenames.map((fname) => {
            const sx =
              fname === viewFile
                ? {
                    color: "var(--tab_cursor_present) !important",
                  }
                : {};
            return fname === filename ? (
              <Tab key={fname} value={fname} ref={selectedEl} sx={sx}>
                {shortnames[fname]}
              </Tab>
            ) : (
              <Tab key={fname} value={fname} sx={sx}>
                {shortnames[fname]}
              </Tab>
            );
          })}
        </TabsList>
      </Tabs>
    </Box>
  );
}

const TabPanel = React.memo(TabPanel_);

function useShortFilenames(filenames: string[]): { [short: string]: string } {
  return useMemo(() => {
    const shortToLong: { [key: string]: string } = {};
    function insert(short: string, long: string) {
      if (shortToLong[short] != null) {
        const existingLong = shortToLong[short];
        const [s1, s2] = disambiguate(existingLong, long);
        if (s2 != null) {
          delete shortToLong[short];
          insert(s1, existingLong);
          insert(s2, long);
        }
      } else {
        shortToLong[short] = long;
      }
    }
    for (const filename of filenames) {
      const last_idx = filename.lastIndexOf("/");
      const basename = filename.slice(last_idx + 1);
      insert(basename, filename);
    }
    const longToShort: { [key: string]: string } = {};
    for (const shortName in shortToLong) {
      longToShort[shortToLong[shortName]] = shortName;
    }
    return longToShort;
  }, [filenames]);
}

function disambiguate(f1: string, f2: string): [string, string | null] {
  if (f1 === f2) {
    return [f1, null];
  }
  let i1 = f1.length;
  let i2 = f2.length;
  let different = false;
  while (!different) {
    i1--;
    i2--;
    while ((i1 > 0 && f1[i1] !== "/") || (i2 > 0 && f2[i2] !== "/")) {
      if (f1[i1] !== f2[i2]) {
        different = true;
      }
      if (i1 > 0 && f1[i1] !== "/") {
        i1--;
      }
      if (i2 > 0 && f2[i2] !== "/") {
        i2--;
      }
    }
  }
  return [f1.slice(i1 + 1), f2.slice(i2 + 1)];
}
