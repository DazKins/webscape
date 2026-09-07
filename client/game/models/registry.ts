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
import { createWoodcuttingAxeModel } from "./definitions/woodcuttingAxe";
import { createFishingRodModel } from "./definitions/fishingRod";
import { createFishingSpotModel } from "./definitions/fishingSpot";
import { createMagicStaffModel } from "./definitions/magicStaff";
import { createWoodenBowModel } from "./definitions/woodenBow";
import type { ModelFactory, ModelInstance, ModelOptions } from "./types";
import {
  createHealthPotionModel, createBreadModel, createAppleModel, createIronOreModel,
  createStoneItemModel, createWoodModel, createLogsModel, createArrowModel,
  createFishModel, createMysteriousKeyModel, createAncientScrollModel,
  createChainmailChestplateModel, createIronLeggingsModel, createLeatherBootsModel,
  createWoodenShieldModel, createUnknownItemModel,
} from "./definitions/items";

export const modelRegistry = {
  human: createHumanModel,
  ironSword: createIronSwordModel,
  woodcuttingAxe: createWoodcuttingAxeModel,
  fishingRod: createFishingRodModel,
  leatherHelmet: createLeatherHelmetModel,
  rat: createRatModel,
  tree: createTreeModel,
  door: createDoorModel,
  chest: createChestModel,
  rock: createRockModel,
  building: createBuildingModel,
  rewarddrop: createRewardDropModel,
  fishingSpot: createFishingSpotModel,
  magicStaff: createMagicStaffModel,
  woodenBow: createWoodenBowModel,
  healthPotion: createHealthPotionModel,
  bread: createBreadModel,
  apple: createAppleModel,
  ironOre: createIronOreModel,
  stone: createStoneItemModel,
  wood: createWoodModel,
  logs: createLogsModel,
  arrow: createArrowModel,
  fish: createFishModel,
  mysteriousKey: createMysteriousKeyModel,
  ancientScroll: createAncientScrollModel,
  chainmailChestplate: createChainmailChestplateModel,
  ironLeggings: createIronLeggingsModel,
  leatherBoots: createLeatherBootsModel,
  woodenShield: createWoodenShieldModel,
  unknownItem: createUnknownItemModel,
} satisfies Record<string, ModelFactory>;

export type ModelName = keyof typeof modelRegistry;

export const modelNames = Object.freeze(Object.keys(modelRegistry) as ModelName[]);

export function isModelName(value: string): value is ModelName {
  return value in modelRegistry;
}

export function createModel(name: ModelName, options: ModelOptions = {}): ModelInstance {
  return modelRegistry[name](options);
}
