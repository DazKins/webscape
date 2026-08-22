import { createBuildingModel } from "./definitions/building";
import { createChestModel } from "./definitions/chest";
import { createDoorModel } from "./definitions/door";
import { createHumanModel } from "./definitions/human";
import { createIronSwordModel } from "./definitions/ironSword";
import { createLeatherHelmetModel } from "./definitions/leatherHelmet";
import { createRatModel } from "./definitions/rat";
import { createRewardDropModel } from "./definitions/rewardDrop";
import { createRockModel } from "./definitions/rock";
import { createTreeModel } from "./definitions/tree";
import type { ModelFactory, ModelInstance, ModelOptions } from "./types";

export const modelRegistry = {
  human: createHumanModel,
  ironSword: createIronSwordModel,
  leatherHelmet: createLeatherHelmetModel,
  rat: createRatModel,
  tree: createTreeModel,
  door: createDoorModel,
  chest: createChestModel,
  rock: createRockModel,
  building: createBuildingModel,
  rewarddrop: createRewardDropModel,
} satisfies Record<string, ModelFactory>;

export type ModelName = keyof typeof modelRegistry;

export const modelNames = Object.freeze(Object.keys(modelRegistry) as ModelName[]);

export function isModelName(value: string): value is ModelName {
  return value in modelRegistry;
}

export function createModel(name: ModelName, options: ModelOptions = {}): ModelInstance {
  return modelRegistry[name](options);
}
