export const ChatMessageEventName = "chatMessage";

export class ChatMessageEvent extends Event {
    constructor(public message: string, public from: string) {
        super(ChatMessageEventName);
        this.message = message;
        this.from = from;
    }
}