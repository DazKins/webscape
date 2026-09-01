import { useEffect, useState } from "react";
import styles from "./inventory.module.css";
import panelStyles from "./uiPanel.module.css";
import Game from "../../game/game";
import { InventoryUpdateEvent, InventoryUpdateEventName } from "../../events/inventory";

type Props = {
  game: Game;
};

const INVENTORY_SLOT_COUNT = 20;

type InventoryItem = {
  id: string;
  name: string;
  type: string;
  renderModel?: string;
  equipmentSlot?: string;
  combatStats?: {
    minDamage: number;
    maxDamage: number;
    accuracyBonus: number;
    armorBonus: number;
    critBonus: number;
    range: number;
    attackSpeedTicks: number;
    attackMethod: string;
    windUpTicks: number;
    travelTicks: number;
    projectileType: string;
  };
};

type ItemIconKind =
  | "weapon"
  | "axe"
  | "fishingRod"
  | "magicStaff"
  | "fish"
  | "helmet"
  | "armor"
  | "boots"
  | "shield"
  | "potion"
  | "food"
  | "ore"
  | "logs"
  | "wood"
  | "stone"
  | "key"
  | "scroll"
  | "default";

type ItemIconDefinition = {
  background: string;
  glow: string;
  shape: string;
};

const itemIcons: Record<ItemIconKind, ItemIconDefinition> = {
  weapon: {
    background: "#2c3342",
    glow: "#a9c7ff",
    shape: `
      <path d="M66 12 78 18 43 55l-8-8 31-35Z" fill="#d9e7f7"/>
      <path d="M37 51 45 59 32 72l-8-8 13-13Z" fill="#7d5a38"/>
      <path d="M31 43 53 65 46 72 24 50l7-7Z" fill="#8aa1bd"/>
      <circle cx="26" cy="70" r="7" fill="#c89a4a"/>
    `,
  },
  axe: {
    background: "#302a24",
    glow: "#d8b17a",
    shape: `
      <path d="M28 78 58 20" stroke="#885a35" stroke-width="9" stroke-linecap="round"/>
      <path d="M50 18c12-4 23-1 31 6L68 46c-7-7-15-11-26-11l8-17Z" fill="#b9c3c8"/>
      <path d="M68 46 81 24" stroke="#eef3f5" stroke-width="5" stroke-linecap="round"/>
      <path d="M25 77 31 65" stroke="#4e3020" stroke-width="11" stroke-linecap="round"/>
    `,
  },
  fishingRod: {
    background: "#25343a",
    glow: "#7ed7e7",
    shape: `
      <path d="M25 80C42 62 57 43 67 17" fill="none" stroke="#9b6a3c" stroke-width="7" stroke-linecap="round"/>
      <path d="M67 17c7 10 11 22 9 35" fill="none" stroke="#d6aa62" stroke-width="4" stroke-linecap="round"/>
      <path d="M76 51c-1 11-8 18-17 23" fill="none" stroke="#d8edf1" stroke-width="2" stroke-linecap="round"/>
      <circle cx="45" cy="57" r="10" fill="none" stroke="#9ba8ae" stroke-width="5"/>
      <circle cx="45" cy="57" r="3" fill="#e4e9eb"/>
      <path d="M22 82 31 71" stroke="#4d3020" stroke-width="10" stroke-linecap="round"/>
    `,
  },
  magicStaff: {
    background: "#253143",
    glow: "#9eeeff",
    shape: `
      <path d="M25 82 66 18" stroke="#6f482b" stroke-width="9" stroke-linecap="round"/>
      <path d="M59 28c8 0 14-5 16-13" fill="none" stroke="#8b5d35" stroke-width="7" stroke-linecap="round"/>
      <path d="M65 19 76 10 84 22 73 33 61 29Z" fill="#c9f4ff"/>
      <circle cx="73" cy="22" r="7" fill="#f1fdff"/>
    `,
  },
  fish: {
    background: "#243844",
    glow: "#73d9ef",
    shape: `
      <path d="M22 50c12-18 34-24 51-8l11-12-2 21 2 21-12-12c-18 14-39 8-50-10Z" fill="#74b9c8"/>
      <path d="M35 43c9-7 22-8 32 0-8 3-13 8-16 17-8-1-14-7-16-17Z" fill="#b8e1e6"/>
      <circle cx="31" cy="46" r="4" fill="#18262c"/>
      <path d="M51 38c3-8 9-13 17-14l-2 17M50 62c4 7 10 10 17 10l-1-15" fill="#5c9eac"/>
    `,
  },
  helmet: {
    background: "#2b3430",
    glow: "#9cd1a4",
    shape: `
      <path d="M22 49c0-18 11-31 27-31s27 13 27 31v10H22V49Z" fill="#8f6b45"/>
      <path d="M30 50h38v11H30V50Z" fill="#4d3a2a"/>
      <path d="M47 19h5v40h-5V19Z" fill="#d6c092"/>
      <path d="M22 58h54v12H22V58Z" fill="#6d4d31"/>
    `,
  },
  armor: {
    background: "#2d3039",
    glow: "#d7dff7",
    shape: `
      <path d="M28 19h40l9 15-8 7v35H27V41l-8-7 9-15Z" fill="#9ea9b6"/>
      <path d="M37 24h22l6 14-8 30H39l-8-30 6-14Z" fill="#4d647c"/>
      <path d="M48 24v44" stroke="#dfe8f2" stroke-width="5"/>
      <path d="M31 41h34" stroke="#26384d" stroke-width="5"/>
    `,
  },
  boots: {
    background: "#342c27",
    glow: "#dfb078",
    shape: `
      <path d="M29 20h17v36l13 8v11H23V64l6-7V20Z" fill="#8a5c39"/>
      <path d="M54 27h14v34l8 6v8H50V63l4-5V27Z" fill="#a06c42"/>
      <path d="M22 71h38v9H22v-9Z" fill="#4b3121"/>
      <path d="M50 71h28v9H50v-9Z" fill="#5c3a24"/>
    `,
  },
  shield: {
    background: "#25363c",
    glow: "#8ed8e0",
    shape: `
      <path d="M48 15 75 25v22c0 18-10 30-27 38-17-8-27-20-27-38V25l27-10Z" fill="#7da6b2"/>
      <path d="M48 22 66 29v17c0 12-6 21-18 28V22Z" fill="#365d67"/>
      <path d="M48 22v52" stroke="#d7eef2" stroke-width="4"/>
    `,
  },
  potion: {
    background: "#302641",
    glow: "#ff8fc7",
    shape: `
      <path d="M40 15h16v13l15 21c8 12 0 29-15 29H40c-15 0-23-17-15-29l15-21V15Z" fill="#d7c9ff"/>
      <path d="M34 51h28c6 10 0 18-7 18H41c-7 0-13-8-7-18Z" fill="#d94d8c"/>
      <path d="M36 13h24v9H36v-9Z" fill="#6e5e86"/>
      <circle cx="56" cy="43" r="4" fill="#ffd8ed"/>
    `,
  },
  food: {
    background: "#3b3023",
    glow: "#f0c36e",
    shape: `
      <path d="M22 55c0-18 14-31 31-31 12 0 22 7 22 19 0 23-23 35-43 30-6-2-10-8-10-18Z" fill="#d79a45"/>
      <path d="M28 54c5 5 21 10 42-7" stroke="#8d5827" stroke-width="7" fill="none" stroke-linecap="round"/>
      <circle cx="47" cy="38" r="4" fill="#f3cd7e"/>
    `,
  },
  ore: {
    background: "#282d35",
    glow: "#aeb8c8",
    shape: `
      <path d="M24 60 34 28l31-9 12 24-16 33H34L24 60Z" fill="#6e7887"/>
      <path d="M34 28 48 49 24 60l10-32Z" fill="#9fa8b4"/>
      <path d="M48 49 65 19l12 24-29 6Z" fill="#d5dde7"/>
      <path d="M48 49 61 76H34L24 60l24-11Z" fill="#4a5260"/>
    `,
  },
  logs: {
    background: "#30271d",
    glow: "#e2a75c",
    shape: `
      <g transform="rotate(-8 48 48)">
        <path d="M27 29h42v22H27Z" fill="#85502e"/>
        <path d="M27 55h42v22H27Z" fill="#6f4127"/>
        <path d="M23 32h38v16H23Z" fill="#9b6236"/>
        <path d="M35 58h38v16H35Z" fill="#89522f"/>
        <ellipse cx="23" cy="40" rx="10" ry="11" fill="#d49a58"/>
        <ellipse cx="35" cy="66" rx="10" ry="11" fill="#c18448"/>
        <ellipse cx="23" cy="40" rx="5" ry="6" fill="none" stroke="#89532f" stroke-width="2"/>
        <ellipse cx="35" cy="66" rx="5" ry="6" fill="none" stroke="#754329" stroke-width="2"/>
        <path d="M24 35c-3 3-3 7 0 10M36 61c-3 3-3 7 0 10" fill="none" stroke="#f0bd76" stroke-width="2" stroke-linecap="round"/>
        <path d="M58 32v16M65 58v16" stroke="#5a321f" stroke-width="3" stroke-linecap="round" opacity="0.7"/>
      </g>
    `,
  },
  wood: {
    background: "#31291f",
    glow: "#d09a55",
    shape: `
      <path d="M26 60 54 20l15 10-28 40-15-10Z" fill="#9c663a"/>
      <path d="M45 74 73 34l10 7-28 40-10-7Z" fill="#7d4e2d"/>
      <path d="M31 58 56 23" stroke="#d8a060" stroke-width="3"/>
      <path d="M51 73 76 38" stroke="#bd814c" stroke-width="3"/>
    `,
  },
  stone: {
    background: "#2e3134",
    glow: "#b6c0c5",
    shape: `
      <path d="M24 59 34 35l24-12 18 17-5 28-24 10-23-19Z" fill="#747d82"/>
      <path d="M34 35 48 57 24 59l10-24Z" fill="#a8b0b4"/>
      <path d="M48 57 58 23l18 17-28 17Z" fill="#879096"/>
      <path d="M48 57 71 68 47 78 24 59l24-2Z" fill="#596168"/>
    `,
  },
  key: {
    background: "#3c3422",
    glow: "#ffd15f",
    shape: `
      <circle cx="35" cy="35" r="16" fill="none" stroke="#f4c24f" stroke-width="9"/>
      <path d="M48 47 75 74" stroke="#f4c24f" stroke-width="9" stroke-linecap="round"/>
      <path d="M64 63 73 54" stroke="#f4c24f" stroke-width="7" stroke-linecap="round"/>
      <path d="M72 71 81 62" stroke="#f4c24f" stroke-width="7" stroke-linecap="round"/>
    `,
  },
  scroll: {
    background: "#383026",
    glow: "#efcf9a",
    shape: `
      <path d="M27 25h34c9 0 12 8 8 15v30H35c-9 0-14-8-9-16V25Z" fill="#dbc08c"/>
      <path d="M29 25c10 0 13 8 8 16H22c-5-8-2-16 7-16Z" fill="#f0d79d"/>
      <path d="M35 55h31" stroke="#8e6e3f" stroke-width="4" stroke-linecap="round"/>
      <path d="M39 42h23" stroke="#8e6e3f" stroke-width="4" stroke-linecap="round"/>
    `,
  },
  default: {
    background: "#2f3138",
    glow: "#b6c2ff",
    shape: `
      <path d="M48 18 72 34v28L48 78 24 62V34l24-16Z" fill="#7986cb"/>
      <path d="M48 18v60M24 34l24 15 24-15" stroke="#d4dbff" stroke-width="5" fill="none"/>
    `,
  },
};

const itemIconSrcCache = new Map<ItemIconKind, string>();

function getItemIconKind(item: InventoryItem): ItemIconKind {
  const name = item.name.toLowerCase();
  if (item.renderModel === "magicStaff") return "magicStaff";
  if (item.type === "axe") return "axe";
  if (item.type === "fishingRod") return "fishingRod";
  if (item.type === "fish") return "fish";

  if (name.includes("key")) return "key";
  if (name.includes("scroll")) return "scroll";
  if (name.includes("potion")) return "potion";
  if (name.includes("bread") || name.includes("apple")) return "food";
  if (name.includes("ore")) return "ore";
  if (name.includes("log")) return "logs";
  if (name.includes("wood")) return "wood";
  if (name.includes("stone")) return "stone";

  switch (item.equipmentSlot) {
    case "head":
      return "helmet";
    case "chest":
    case "legs":
      return "armor";
    case "feet":
      return "boots";
    case "offhand":
      return "shield";
    case "weapon":
      return "weapon";
  }

  switch (item.type) {
    case "weapon":
      return "weapon";
    case "armor":
      return "armor";
    case "shield":
      return "shield";
    case "consumable":
      return "potion";
    case "material":
      return "ore";
    case "quest":
      return "key";
    default:
      return "default";
  }
}

function getItemIconSrc(item: InventoryItem): string {
  const kind = getItemIconKind(item);
  const cached = itemIconSrcCache.get(kind);
  if (cached) {
    return cached;
  }

  const icon = itemIcons[kind];
  const svg = `
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">
      <rect width="96" height="96" rx="12" fill="${icon.background}"/>
      <circle cx="70" cy="25" r="22" fill="${icon.glow}" opacity="0.22"/>
      <circle cx="26" cy="76" r="16" fill="#000000" opacity="0.18"/>
      ${icon.shape}
    </svg>
  `;
  const src = `data:image/svg+xml,${encodeURIComponent(svg)}`;
  itemIconSrcCache.set(kind, src);
  return src;
}

function getItemTitle(item: InventoryItem): string {
  return `${item.name} (${item.type})${
    item.equipmentSlot ? ` - Equipable: ${item.equipmentSlot}` : ""
  }${
    item.combatStats
      ? ` - Dmg ${item.combatStats.minDamage}-${item.combatStats.maxDamage}, Acc ${item.combatStats.accuracyBonus}, Armor ${item.combatStats.armorBonus}`
      : ""
  }`;
}

function useInventoryState(props: Props) {
  const [items, setItems] = useState<InventoryItem[]>([]);
  const [equippedSlots, setEquippedSlots] = useState<Record<string, InventoryItem | null>>({});

  const updateInventory = () => {
    const myEntity = props.game.getMyEntity();
    if (!myEntity) {
      setItems([]);
      setEquippedSlots({});
      return;
    }

    const inventoryComponent = myEntity.getComponent("inventory");
    if (!inventoryComponent || !inventoryComponent.items) {
      setItems([]);
      setEquippedSlots({});
      return;
    }

    setItems(inventoryComponent.items || []);

    const equippedComponent = myEntity.getComponent("equipped");
    if (equippedComponent && equippedComponent.slots) {
      setEquippedSlots(equippedComponent.slots);
    } else {
      setEquippedSlots({});
    }
  };

  useEffect(() => {
    // Update inventory initially
    updateInventory();

    // Listen for inventory update events
    const handler = (_event: InventoryUpdateEvent) => {
      updateInventory();
    };

    props.game.addEventListener(InventoryUpdateEventName, handler as EventListener);
    return () => {
      props.game.removeEventListener(InventoryUpdateEventName, handler as EventListener);
    };
  }, [props.game]);

  return { items, equippedSlots };
}

export function InventoryBackpackContent(props: Props) {
  const { items } = useInventoryState(props);
  const handleEquip = (itemId: string) => {
    props.game.handleEquipItem(itemId);
  };

  const backpackItems = items.slice(0, INVENTORY_SLOT_COUNT);
  const backpackSlots = Array.from({ length: INVENTORY_SLOT_COUNT }, (_, index) => ({
    index,
    item: backpackItems[index],
  }));

  return (
    <div className={`${panelStyles.panelContent} ${styles.content}`}>
      <div className={styles.sectionHeader}>
        Backpack {backpackItems.length}/{INVENTORY_SLOT_COUNT}
      </div>
      <div className={styles.itemsGrid}>
        {backpackSlots.map(({ index, item }) =>
          item ? (
            <div
              key={item.id}
              className={`${styles.item} ${item.equipmentSlot ? styles.equipable : ""}`}
              title={getItemTitle(item)}
            >
              <img
                className={styles.itemImage}
                src={getItemIconSrc(item)}
                alt={item.name}
                draggable={false}
              />
              {item.equipmentSlot && (
                <button
                  className={styles.equipButton}
                  onClick={() => handleEquip(item.id)}
                >
                  Equip
                </button>
              )}
            </div>
          ) : (
            <div
              key={`empty-${index}`}
              className={`${styles.item} ${styles.emptySlot}`}
              aria-label={`Empty inventory slot ${index + 1}`}
            />
          )
        )}
      </div>
    </div>
  );
}

export function EquipmentContent(props: Props) {
  const { equippedSlots } = useInventoryState(props);
  const handleUnequip = (slot: string) => {
    props.game.handleUnequipSlot(slot);
  };

  const equippedEntries = Object.entries(equippedSlots);

  return (
    <div className={panelStyles.panelContent}>
      <div className={styles.sectionHeader}>Equipment</div>
      {equippedEntries.length === 0 ? (
        <div className={styles.emptyMessage}>No equipment</div>
      ) : (
        <div className={styles.equippedList}>
          {equippedEntries.map(([slot, item]) => (
            <div key={slot} className={styles.equippedItem}>
              <div className={styles.equippedSlot}>{slot}</div>
              {item ? (
                <div className={styles.equippedDetails}>
                  <img
                    className={styles.equippedImage}
                    src={getItemIconSrc(item)}
                    alt={item.name}
                    title={getItemTitle(item)}
                    draggable={false}
                  />
                  <button
                    className={styles.equipButton}
                    onClick={() => handleUnequip(slot)}
                  >
                    Unequip
                  </button>
                </div>
              ) : (
                <div className={styles.equippedEmpty}>Empty</div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default function Inventory(props: Props) {
  return (
    <div className={`${panelStyles.panel} ${styles.container}`}>
      <div className={panelStyles.panelHeader}>Inventory</div>
      <InventoryBackpackContent game={props.game} />
    </div>
  );
}
