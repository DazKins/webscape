declare const __BUILD_REVISION__: string;
declare const __BUILD_DIRTY__: boolean;

declare module "*.module.css";

interface Window {
  __MODEL_LAB_READY__?: boolean;
  __MODEL_LAB_MODELS__?: string[];
}
