export default class ChatBox {
  box: HTMLDivElement;

  constructor() {
    this.box = document.createElement("div");
    this.box.style.position = "absolute";
    this.box.style.display = "none";
    this.box.style.left = "0px";
    this.box.style.top = "0px";

    this.box.addEventListener("click", (event) => {
      event.stopPropagation();
    });

    document.body.appendChild(this.box);

    this.update("");
  }

  update(typedText: string) {
    this.box.innerHTML = `
      <div class="chat-box">
        <div class="chat-box-header">
          <div class="chat-box-header-title">Chat</div>
        </div>
        <div class="chat-box-content">
        </div>
        <div class="chat-box-input">
        > ${typedText}*
        </div>
      </div>
    `;
  }

  show() {
    this.box.style.display = "block";
  }

  hide() {
    this.box.style.display = "none";
  }
}
