import { useEffect, useState } from "react";
import { ModelController } from "../../core/model";
import "@pages/panel/Panel.css";
import { InitProgressReport } from "@mlc-ai/web-llm";

export default function Panel() {
  const [loaded, setLoaded] = useState(false);
  let model: any = null;

  function callback(progress: InitProgressReport) {
    console.log("Qwen2-0.5B-Instruct-q4f16_1-MLC", progress);
    setLoaded(true);
    model = ModelController.getInstance().getModel(
      "Qwen2-0.5B-Instruct-q4f16_1-MLC",
    );
  }

  useEffect(() => {
    const modelController = ModelController.getInstance();
    modelController.registerModel("Qwen2-0.5B-Instruct-q4f16_1-MLC", callback);
  });

  return (
    <div className="container">
      <h1>Side Panel AI Model Status: {loaded ? "Loaded" : "Loading"}</h1>
      <div>
        <button
          onClick={async () => {
            if (model) {
              const text = await model.chat.completions.create({
                stream: false,
                messages: [
                  {
                    role: "user",
                    content: "poop",
                  },
                ],
              });
              console.log(text);
            }
          }}
        >
          Reset
        </button>
      </div>
    </div>
  );
}
