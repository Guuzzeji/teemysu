import {
  createWebLLMServiceWorker,
  ServiceWorkerState,
} from "../../core/model";

console.log("background script loaded");

const state: ServiceWorkerState = {
  handler: null,
};

createWebLLMServiceWorker(state);
