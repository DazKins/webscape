export const createMessage = (type, data) => {
  return {
    metadata: {
      type,
      time: new Date().toISOString(),
    },
    data: data,
  };
};

export const getMessageType = (msg) => {
  return msg.metadata.type;
};

export const getMessageData = (msg) => {
  return msg.data;
};
