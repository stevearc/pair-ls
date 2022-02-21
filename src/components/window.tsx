import * as React from "react";
import styled from "@emotion/styled";
import { AppContext, View, Range, File } from "../state";
import { type HLJSApi } from "highlight.js";
const { useContext, useEffect, useMemo, useRef } = React;

let _hljs: HLJSApi | null = null;
function lazyImport(): HLJSApi {
  if (_hljs != null) {
    return _hljs;
  }
  throw import("highlight.js").then((mod) => {
    _hljs = mod.default;
  });
}

const Container = styled.div`
  display: flex;
  flex-direction: row;
  overflow: auto;
`;
const Gutter = styled.pre`
  display: flex;
  flex-direction: column;
  flex-grow: 0;
  margin: 0;
  /* TODO I don't know why this doesn't align  at 0.5rem */
  padding: 0.51rem 0;
  text-align: right;
  color: var(--fg_gutter);
`;

export default function WindowComponent({ file_id }: { file_id: number }) {
  const { client, state } = useContext(AppContext);
  const file = state.files[file_id];
  useEffect(() => {
    if (client != null && file != null) {
      client.getText(file.filename);
    }
  }, [file_id, client]);
  if (client == null) {
    return null;
  }
  if (file == null) {
    return null;
  }
  const lines = file.lines;
  if (lines == null) {
    throw client.getText(file.filename);
  }
  const hljs = lazyImport();
  return (
    <Window hljs={hljs} file={file} view={state.view} follow={state.follow} />
  );
}

function Window_({
  file,
  follow,
  view,
  hljs,
}: {
  file: File;
  follow: boolean;
  view?: View | null;
  hljs: HLJSApi;
}) {
  const cursorRef = useRef<HTMLDivElement | null>(null);
  const prevFile = usePrevious(file.filename);
  useEffect(() => {
    if (cursorRef.current != null && follow) {
      const scrollBehavior = prevFile === file.filename ? "smooth" : "auto";
      cursorRef.current.scrollIntoView({
        behavior: scrollBehavior,
        block: "center",
        inline: "end",
      });
    }
  }, [view]);
  const code = useMemo(
    () => hljs.highlightAuto(file.lines!.join("\n"), langToHLJS(file.language)),
    [file]
  );
  return (
    <Container>
      <Gutter>
        <code>{file.lines!.map((_, i) => `${i + 1}`).join("\n")}</code>
      </Gutter>
      <pre className="hljs">
        {view != null && view?.file_id === file.id && (
          <React.Fragment>
            <Cursor ref={cursorRef} lines={file.lines!} view={view}></Cursor>
            {view.range != null && (
              <Selection lines={file.lines!} range={view.range}></Selection>
            )}
          </React.Fragment>
        )}
        <Code language={code.language} markup={code.value} />
      </pre>
    </Container>
  );
}

const Window = React.memo(Window_);

function Code_({
  language,
  markup,
}: {
  language: string | undefined;
  markup: string;
}) {
  return (
    <code
      className={`${language}`}
      dangerouslySetInnerHTML={{ __html: markup }}
    />
  );
}
const Code = React.memo(Code_);

const Cursor = React.forwardRef<
  HTMLDivElement,
  { view: View; lines: string[] }
>(({ view, lines }, ref) => {
  if (view.line >= lines.length) {
    return null;
  }
  const tabOffset = calcTabOffset(lines, view.line, view.character);
  const CursorDiv = styled.div`
    position: absolute;
    top: ${0.5 + 1.2 * 0.8 * (view?.line ?? 0)}rem;
    left: calc(0.5rem + ${tabOffset + (view?.character ?? 0)}ch);
    font-family: monospace;
    font-size: 0.8rem;
    display: inline-block;
    height: 1rem;
    line-height: 1.2;
    mix-blend-mode: difference;
    width: 1ch;
    background: var(--cursor);
    opacity: 1;
    animation: blink 0.5s linear infinite alternate;
  `;

  return <CursorDiv ref={ref}></CursorDiv>;
});

function SelectionDiv({
  lines,
  line,
  start,
  end,
}: {
  lines: string[];
  line: number;
  start: number;
  end: number;
}) {
  if (line >= lines.length) {
    return null;
  }
  if (lines[line] === "") {
    end = 1;
  }
  start += calcTabOffset(lines, line, start, false);
  end += calcTabOffset(lines, line, end);
  let width = `${end - start}ch`;
  if (end === -1) {
    width = start === 0 ? "100%" : `calc(100% - ${start}ch)`;
  }
  const SDiv = styled.div`
    position: absolute;
    top: ${0.5 + 1.2 * 0.8 * line}rem;
    left: calc(0.5rem + ${start}ch);
    font-family: monospace;
    font-size: 0.8rem;
    display: inline-block;
    height: 1rem;
    line-height: 1.2;
    mix-blend-mode: difference;
    width: ${width};
    background: var(--selection);
  `;
  return <SDiv></SDiv>;
}

function Selection({ lines, range }: { lines: string[]; range: Range }) {
  let start, end;
  if (
    range.end.line < range.start.line ||
    (range.end.line == range.start.line &&
      range.end.character < range.start.character)
  ) {
    start = range.end;
    end = range.start;
  } else {
    end = range.end;
    start = range.start;
  }
  const elements = [];
  for (let i = start.line; i <= end.line; i++) {
    const start_col = start.line === i ? start.character : 0;
    const end_col = end.line === i ? end.character : -1;
    elements.push(
      <SelectionDiv
        key={i}
        lines={lines}
        line={i}
        start={start_col}
        end={end_col}
      />
    );
  }
  return <React.Fragment>{elements}</React.Fragment>;
}

function calcTabOffset(
  lines: string[],
  line: number,
  character: number,
  position_end: boolean = true
): number {
  const text = lines[line];
  let numTabs = 0;
  const end = position_end ? character + 1 : character;
  for (let i = 0; i < text.length && i < end; i++) {
    if (text[i] === "\t") {
      numTabs++;
    }
  }
  return numTabs;
}
function usePrevious<T>(value: T, update: boolean = true): T | null {
  const ref = useRef<T | null>(null);
  useEffect(() => {
    if (update) {
      ref.current = value;
    }
  }, [value]);
  return ref.current;
}

const LSP_TO_HLJS_LANG: { [language: string]: string } = {
  javascriptreact: "jsx",
  ["javascript.jsx"]: "jsx",
  typescriptreact: "typescript",
  ["typescript.jsx"]: "typescript",
};

/**
 * Map to supported HLJS languages
 * see https://github.com/highlightjs/highlight.js/blob/main/SUPPORTED_LANGUAGES.md
 */
function langToHLJS(language: string): string[] {
  if (LSP_TO_HLJS_LANG[language] != null) {
    return [LSP_TO_HLJS_LANG[language]];
  } else {
    return [language];
  }
}
