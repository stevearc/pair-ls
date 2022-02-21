import * as React from "react";
import { type AlertProps } from "@mui/material/Alert";
import BaseClient from "./base_client";
import {
  colorschemes,
  setColorScheme,
  ColorScheme,
} from "./colors/colorschemes";

export type Range = {
  start: {
    line: number;
    character: number;
  };
  end: {
    line: number;
    character: number;
  };
};

export type View = {
  file_id: number;
  line: number;
  character: number;
  range?: Range;
};

export type ChangeTextRange = {
  start_line: number;
  end_line: number;
  text: string[];
};

export type File = {
  filename: string;
  id: number;
  language: string;
  lines?: string[];
};

export type FileMap = {
  [file_id: number]: File;
};

export type SyncResponse = {
  files: File[];
  view?: View | null;
};

export type AlertWrapper = {
  id: number;
  text: string | (() => string);
  props: AlertProps;
};

export type AppState = {
  file_id?: number;
  colorscheme: ColorScheme;
  view?: View | null;
  follow: boolean;
  files: FileMap;
  alerts: AlertWrapper[];
};

export type AppAction =
  // Actions from the server
  | {
      type: "initialize";
      sync: SyncResponse;
    }
  | {
      type: "openFile";
      filename: string;
      id: number;
      language: string;
    }
  | {
      type: "closeFile";
      file_id: number;
    }
  | {
      type: "setText";
      file_id: number;
      text: string[];
    }
  | {
      type: "updateView";
      view: View;
    }
  | {
      type: "updateText";
      file_id: number;
      changes: ChangeTextRange[];
    }
  // User actions
  | {
      type: "toggleFollow";
    }
  | {
      type: "selectFile";
      file_id: number;
    }
  | {
      type: "setColorScheme";
      colorscheme: ColorScheme;
    }
  // Responses to events
  | {
      type: "toast";
      id: number;
      props: AlertProps;
      text: string | (() => string);
    }
  | {
      type: "removeToast";
      id: number;
    };

export type Dispatcher = React.Dispatch<AppAction>;

export function reducer(state: AppState, action: AppAction): AppState {
  switch (action.type) {
    case "initialize": {
      const files: FileMap = {};
      for (const file of action.sync.files) {
        files[file.id] = file;
      }
      return {
        ...state,
        files,
        file_id: action.sync.view?.file_id ?? action.sync.files[0]?.id,
        view: action.sync.view,
      };
    }
    case "openFile":
      return {
        ...state,
        file_id: state.file_id ?? action.id,
        files: {
          ...state.files,
          [action.id]: {
            filename: action.filename,
            id: action.id,
            language: action.language,
          },
        },
      };
    case "closeFile":
      if (state.files[action.file_id] == null) {
        return state;
      }
      const newFiles = { ...state.files };
      delete newFiles[action.file_id];
      return {
        ...state,
        files: newFiles,
      };
    case "setText":
      return {
        ...state,
        files: {
          ...state.files,
          [action.file_id]: {
            ...state.files[action.file_id],
            lines: action.text,
          },
        },
      };
    case "updateText": {
      const newLines = [...(state.files[action.file_id].lines ?? [])];
      for (const change of action.changes) {
        newLines.splice(
          change.start_line,
          change.end_line - change.start_line + 1,
          ...change.text
        );
      }
      return {
        ...state,
        files: {
          ...state.files,
          [action.file_id]: {
            ...state.files[action.file_id],
            lines: newLines,
          },
        },
      };
    }
    case "updateView":
      if (state.follow) {
        return {
          ...state,
          file_id: action.view.file_id,
          view: action.view,
        };
      } else {
        return {
          ...state,
          view: action.view,
        };
      }
    case "selectFile":
      return {
        ...state,
        follow: false,
        file_id: action.file_id,
      };
    case "toggleFollow":
      const newFollow = !state.follow;
      if (newFollow) {
        return {
          ...state,
          file_id: state.view?.file_id ?? state.file_id,
          follow: newFollow,
        };
      } else {
        return {
          ...state,
          follow: newFollow,
        };
      }
    case "setColorScheme":
      localStorage.setItem("colorscheme", action.colorscheme);
      if (state.colorscheme !== action.colorscheme) {
        setColorScheme(action.colorscheme);
        return {
          ...state,
          colorscheme: action.colorscheme,
        };
      } else {
        return state;
      }
    case "toast": {
      return {
        ...state,
        alerts: [...state.alerts, action],
      };
    }
    case "removeToast":
      return {
        ...state,
        alerts: state.alerts.filter((a) => a.id !== action.id),
      };
    default:
      const v: never = action;
      console.error("Unknown action", v);
      return state;
  }
}

let alertID = 0;
function createAlertID(): number {
  return alertID++;
}

export function showToast(
  dispatch: Dispatcher,
  text: string | (() => string),
  props: AlertProps,
  duration: number | null = 4000
): number {
  const id = createAlertID();
  dispatch({
    type: "toast",
    id,
    text,
    props,
  });
  if (duration != null) {
    setTimeout(() => {
      dispatch({ type: "removeToast", id });
    }, duration);
  }

  return id;
}

export function getInitialState(): AppState {
  let colorscheme = localStorage.getItem("colorscheme") as
    | ColorScheme
    | undefined;
  if (colorscheme == null || !colorschemes.hasOwnProperty(colorscheme)) {
    const darkTheme = window.matchMedia("(prefers-color-scheme: dark)").matches;
    colorscheme = darkTheme ? "tokyonight" : "solarized_light";
  }
  setColorScheme(colorscheme);
  return {
    colorscheme,
    file_id: undefined,
    files: {},
    follow: true,
    view: undefined,
    alerts: [],
  };
}

export type AppContextType = {
  state: AppState;
  dispatch: Dispatcher;
  client: BaseClient | null;
  setClient: (client: BaseClient | null) => void;
};
export const AppContext = React.createContext<AppContextType>({
  state: getInitialState(),
  dispatch: () => {
    console.error("AppContext is uninitialized");
  },
  client: null,
  setClient: () => {
    console.error("AppContext is uninitialized");
  },
});
