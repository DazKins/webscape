export type Command = {
  type: string;
  data: any;
};

export const createCommand = (type: string, data: any): Command => {
  return {
    type,
    data: data,
  };
};

export const getCommandType = (cmd: Command): string => {
  return cmd.type;
};

export const getCommandData = (cmd: Command): any => {
  return cmd.data;
};
