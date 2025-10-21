export type Command = {
  type: string;
  data: any;
};

export const createCommand = (type: string, data: any): Command => {
  return {
    type,
    data,
  };
};
