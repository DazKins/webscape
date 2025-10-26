import { CSS2DObject } from "three/examples/jsm/renderers/CSS2DRenderer.js";
import { createRoot, Root } from "react-dom/client";
import React from "react";

export function createReactCss2dObject<P>(
  Component: React.ComponentType<P>,
  props: P
): CSS2DObject {
  const container = document.createElement("div");
  container.style.pointerEvents = "none";

  const root: Root = createRoot(container);
  root.render(React.createElement(Component as React.ElementType, props));

  return new CSS2DObject(container);
}
