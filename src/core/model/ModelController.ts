import {
  ChatCompletionMessageParam,
  CreateExtensionServiceWorkerMLCEngine,
  MLCEngineInterface,
  InitProgressReport,
} from "@mlc-ai/web-llm";

export class ModelController {
  private modelMap: Map<string, MLCEngineInterface> = new Map<
    string,
    MLCEngineInterface
  >();

  static instance: ModelController;
  static getInstance() {
    if (!ModelController.instance) {
      ModelController.instance = new ModelController();
    }
    return ModelController.instance;
  }

  async registerModel(
    modelName: string,
    initProgressCallback?: (progress: InitProgressReport) => void,
  ) {
    const model: MLCEngineInterface =
      await CreateExtensionServiceWorkerMLCEngine(modelName, {
        initProgressCallback: initProgressCallback,
      });
    this.modelMap.set(modelName, model);
  }

  getModel(modelName: string) {
    return this.modelMap.get(modelName);
  }
}
