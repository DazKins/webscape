import { AssetCache } from "./assetCache";
import { setTreeDamageStage } from "./definitions/tree";
import {
  createLanternModel, createBenchModel, createTavernTableModel,
  createBookcaseModel, createShopCounterModel, createArcheryTargetModel,
  createFountainModel,
} from "./definitions/villageScenery";
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
  createGoldModel, createHealthPotionModel, createBreadModel, createAppleModel, createIronOreModel,
  createStoneItemModel, createWoodModel, createLogsModel, createArrowModel,
  createFishModel, createMysteriousKeyModel, createAncientScrollModel,
  createChainmailChestplateModel, createIronLeggingsModel, createLeatherBootsModel,
  createWoodenShieldModel, createUnknownItemModel,
} from "./definitions/items";

export const modelRegistry = {
  lantern: createLanternModel,
  bench: createBenchModel,
  tavernTable: createTavernTableModel,
  bookcase: createBookcaseModel,
  shopCounter: createShopCounterModel,
  archeryTarget: createArcheryTargetModel,
  fountain: createFountainModel,
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
  gold: createGoldModel,
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
  // Share human parts without caching every combination of appearance options.
  if (name === "human") return modelRegistry[name](options);
  const { damageStage: _damageStage, ...visualOptions } = options;
  const key = JSON.stringify([name, Object.entries(visualOptions).sort(([a], [b]) => a.localeCompare(b))]);
  let template = templates.get(key);
  if (!template) {
    template = modelRegistry[name](visualOptions);
    templates.set(key, template);
  }
  const instance = template.clone();
  if (name === "tree") setTreeDamageStage(instance.root, options.damageStage ?? 0);
  return instance;
}

const templates = new AssetCache<ModelInstance>(64, model => model.dispose());

export const modelAssets = { instantiate: createModel, clear: () => templates.clear() };
