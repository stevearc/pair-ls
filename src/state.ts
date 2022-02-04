import * as React from "react";
import { type AlertProps } from "@mui/material/Alert";
import Client from "./client";
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
  filename: string;
  line: number;
  character: number;
  range?: Range;
};

export type File = {
  filename: string;
  language: string;
  lines?: string[];
};

export type FileMap = {
  [filename: string]: File;
};

export type SyncResponse = {
  files: FileMap;
  view?: View | null;
};

export type AlertWrapper = {
  id: number;
  text: string | (() => string);
  props: AlertProps;
};

export type AppState = {
  filename?: string;
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
      language: string;
    }
  | {
      type: "closeFile";
      filename: string;
    }
  | {
      type: "setText";
      filename: string;
      text: string[];
    }
  | {
      type: "updateView";
      view: View;
    }
  // User actions
  | {
      type: "toggleFollow";
    }
  | {
      type: "selectFile";
      filename: string;
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
    case "initialize":
      return {
        ...state,
        files: action.sync.files,
        filename:
          action.sync.view?.filename ?? Object.keys(action.sync.files)[0],
        view: action.sync.view,
      };
    case "openFile":
      return {
        ...state,
        filename: state.filename ?? action.filename,
        files: {
          ...state.files,
          [action.filename]: {
            filename: action.filename,
            language: action.language,
          },
        },
      };
    case "closeFile":
      if (state.files[action.filename] == null) {
        return state;
      }
      const newFiles = { ...state.files };
      delete newFiles[action.filename];
      return {
        ...state,
        files: newFiles,
      };
    case "setText":
      return {
        ...state,
        files: {
          ...state.files,
          [action.filename]: {
            ...state.files[action.filename],
            lines: action.text,
          },
        },
      };
    case "updateView":
      if (state.follow) {
        return {
          ...state,
          filename: action.view.filename,
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
        filename: action.filename,
      };
    case "toggleFollow":
      const newFollow = !state.follow;
      if (newFollow) {
        return {
          ...state,
          filename: state.view?.filename ?? state.filename,
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
  duration: number | null = 4
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
    filename: undefined,
    files: {},
    follow: true,
    view: undefined,
    alerts: [],
  };
}

export type AppContextType = {
  state: AppState;
  dispatch: Dispatcher;
  client: Client | null;
  setClient: (client: Client | null) => void;
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
