import { ExtensionServiceWorkerMLCEngineHandler } from "@mlc-ai/web-llm";
import browser from "webextension-polyfill";

export type ServiceWorkerState = {
  handler: ExtensionServiceWorkerMLCEngineHandler | null;
};

// Pass a state object so the caller can see updates
export function createWebLLMServiceWorker(state: ServiceWorkerState) {
  browser.runtime.onConnect.addListener((port: any) => {
    console.log("onConnect", port);
    if (state.handler === null) {
      // Reassigning a property on the object updates it everywhere
      state.handler = new ExtensionServiceWorkerMLCEngineHandler(port);
    } else {
      state.handler.setPort(port);
    }

    port.onMessage.addListener(state.handler.onmessage.bind(state.handler));
  });
}
