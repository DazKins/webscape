import { CSS2DObject } from "three/examples/jsm/Addons.js";

export const createEntityChat = (text: string): CSS2DObject => {
  const box = document.createElement("div");
  box.style.position = "absolute";
  box.style.display = "block";
  box.innerHTML = `<p>${text}</p>`;

  const css2dObject = new CSS2DObject(box);
  css2dObject.position.set(0, 2, 0);

  return css2dObject;
};
