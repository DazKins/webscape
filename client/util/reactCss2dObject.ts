import { CSS2DObject } from "three/examples/jsm/renderers/CSS2DRenderer.js";
import { createRoot, Root } from "react-dom/client";
import React from "react";

export function createReactCss2dObject<P>(
  Component: React.ComponentType<P>,
  props: P
): CSS2DObject {
  const outer = document.createElement("div");
  outer.style.position = "absolute";
  outer.style.top = "0";
  outer.style.left = "0";
  outer.style.pointerEvents = "none";

  const inner = document.createElement("div");
  outer.appendChild(inner);

  const root: Root = createRoot(inner);
  root.render(React.createElement(Component as React.ElementType, props));

  return new CSS2DObject(outer);
}
