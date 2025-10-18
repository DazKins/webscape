export type Message = {
  metadata: {
    type: string;
    time: string;
  };
  data: any;
}

export const createMessage = (type: string, data: any): Message => {
  return {
    metadata: {
      type,
      time: new Date().toISOString(),
    },
    data: data,
  };
};

export const getMessageType = (msg: Message): string => {
  return msg.metadata.type;
};

export const getMessageData = (msg: Message): any => {
  return msg.data;
};
