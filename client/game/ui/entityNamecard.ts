import { CSS2DObject } from "three/examples/jsm/Addons.js";

export const createEntityNamecard = (
  name: string,
  positionX: number,
  positionY: number,
  positionZ: number
): CSS2DObject => {
  const box = document.createElement("div");
  box.style.position = "absolute";
  box.style.display = "block";
  box.innerHTML = `
    <div class="entity-namecard">
      ${name}
    </div>
  `;

  const css2dObject = new CSS2DObject(box);
  css2dObject.position.set(positionX, positionY, positionZ);

  return css2dObject;
};
