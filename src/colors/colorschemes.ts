import * as React from "react";
import { createTheme } from "@mui/material/styles";
import { type PaletteMode } from "@mui/material";
import tokyonight from "./tokyonight";
import solarized_light from "./solarized_light";
const { useMemo } = React;

const colormap = {
  tokyonight,
  solarized_light,
};

export type ColorScheme = keyof typeof colormap;

export const colorschemes: {
  [key in ColorScheme]: string;
} = {
  tokyonight: "Tokyo Night",
  solarized_light: "Solarized Light",
};

const colorModes: {
  [key in ColorScheme]: PaletteMode;
} = {
  tokyonight: "dark",
  solarized_light: "light",
};

export function setColorScheme(cs: ColorScheme) {
  const scheme = colormap[cs];
  const root = document.querySelector(":root");
  for (const type of colorGroups) {
    // @ts-ignore
    let color = scheme[type];
    let fallback = type;
    while (color == null && fallback != null) {
      // @ts-ignore
      fallback = colorFallbacks[fallback];
      // @ts-ignore
      color = scheme[fallback];
    }

    // @ts-ignore
    root.style.setProperty(`--${type}`, color == null ? "" : color);
  }
}

const colorFallbacks = {
  bg: null, // Code background
  paper_bg: "bg", // Background for MUI "paper" surfaces (e.g. dialogs)
  fg: null, // Code text
  fg_gutter: "fg", // Line numbers
  cursor: "fg",
  selection: "cursor", // Selected text
  tab_hover_bg: "selection",
  tab_hover_fg: "fg",
  tab_selected_bg: "tab_hover_bg",
  tab_selected_fg: "tab_hover_fg",
  tab_cursor_present: "keyword",
  success: null,
  // from Treesitter
  preproc: null, // Preprocessor directives
  annotation: "preproc",
  attribute: "preproc",
  boolean: "constant",
  character: "constant",
  comment: null,
  note: "comment",
  warning: "identifier",
  danger: "keyword",
  constructor: "keywordFunction",
  conditional: "repeat",
  constant: null,
  constBuiltin: "keyword", // Built-in constant values: `nil` in Lua.
  constMacro: "preproc",
  error: "comment", // Syntax/parser errors
  exception: "keyword",
  field: "symbol",
  float: "number",
  function: null,
  funcBuiltin: "keyword",
  funcMacro: "preproc",
  include: "preproc",
  keyword: null,
  keywordFunction: "keyword",
  keywordOperator: "keyword",
  keywordReturn: "keyword",
  label: "keyword",
  method: "function",
  namespace: "include",
  none: "fg",
  number: "constant",
  operator: null,
  parameter: "symbol",
  parameterReference: "parameter",
  property: "field",
  punctDelimiter: null, // Punctuation delimiters: Periods, commas, semicolons, etc.
  punctBracket: "punctDelimiter",
  punctSpecial: "punctDelimiter",
  repeat: "keyword", // Keywords related to loops: `for`, `while`, etc.
  string: null,
  stringRegex: "string",
  stringSpecial: "symbol",
  stringEscape: "symbol",
  symbol: null,
  type: null,
  typeBuiltin: "type",
  variable: "fg",
  variableBuiltin: "keyword",
  tag: "keyword",
  tagDelimiter: "punctDelimiter",
  text: "none",
  textReference: "constant",
  // emphasis: null,
  // underline: null,
  // strike: null,
  title: "none",
  literal: "string",
  // URI: null,
};

const colorGroups = Object.keys(colorFallbacks);

function cssVar(name: string): string {
  const color = getComputedStyle(document.documentElement)
    .getPropertyValue(name)
    .trim();
  if (color === "") {
    console.warn("No color for", name);
    return "#f00";
  }
  return color;
}

export function useTheme(colorscheme: ColorScheme) {
  return useMemo(() => {
    return createTheme({
      palette: {
        mode: colorModes[colorscheme],
        background: {
          default: cssVar("--bg"),
          paper: cssVar("--paper_bg"),
        },
        primary: {
          main: cssVar("--fg"),
          contrastText: cssVar("--bg"),
        },
        secondary: {
          main: cssVar("--keyword"),
        },
        error: {
          main: cssVar("--danger"),
        },
        warning: {
          main: cssVar("--warning"),
        },
        info: {
          main: cssVar("--note"),
        },
        success: {
          main: cssVar("--success"),
        },
      },
    });
  }, [colorscheme]);
}
