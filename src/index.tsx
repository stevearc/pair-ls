import * as ReactDOM from "react-dom";
import * as React from "react";
import App from "./app";
import SimpleRoot from "./components/simple_root";
import RtcRoot from "./components/rtc_root";
import adapter from "webrtc-adapter";

(function () {
  const domContainer = document.querySelector("#root");
  const attr = domContainer?.getAttribute("rtc") ?? "";
  const isRTC = attr === "true" || process.env.IS_STATIC;
  const Content = isRTC ? RtcRoot : SimpleRoot;

  // We don't really need this, but I'm using it to make sure we don't lose the
  // adapter shims due to tree shaking (it relies on import side effects)
  if (adapter.browserDetails.browser == null) {
    console.log("Unknown browser");
  }
  ReactDOM.render(
    <App>
      <Content />
    </App>,
    domContainer
  );
})();
