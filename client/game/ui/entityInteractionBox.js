class EntityInteractionBox {
  constructor() {
    this.box = document.createElement("div");
    this.box.style.position = "absolute";
    this.box.style.display = "none";

    this.box.addEventListener("click", (event) => {
      event.stopPropagation();
    });

    document.body.appendChild(this.box);
  }

  show(entity, positionX, positionY, name, interactionOptions) {
    this.box.style.left = `${positionX}px`;
    this.box.style.top = `${positionY}px`;
    this.box.style.display = "block";
    this.box.innerHTML = `
      <div class="entity-interaction-box">
        <div class="title">${name || "Unnamed?!?!"}</div>
        <div class="interaction-options">
          ${interactionOptions
            .map((option) => `<div class="interaction-option">${option}</div>`)
            .join("")}
        </div>
      </div>
    `;
  }

  hide() {
    this.box.style.display = "none";
  }
}

export default EntityInteractionBox;
