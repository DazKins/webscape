import { useEffect, useState } from "react";
import styles from "./inventory.module.css";
import panelStyles from "./uiPanel.module.css";
import Game from "../../game/game";
import { InventoryUpdateEvent, InventoryUpdateEventName } from "../../events/inventory";

type Props = {
  game: Game;
};

type InventoryItem = {
  id: string;
  name: string;
  type: string;
  equipmentSlot?: string;
  combatStats?: {
    minDamage: number;
    maxDamage: number;
    accuracyBonus: number;
    armorBonus: number;
    critBonus: number;
    range: number;
    attackSpeedTicks: number;
  };
};

export default function Inventory(props: Props) {
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

  const myEntity = props.game.getMyEntity();
  if (!myEntity) {
    return null;
  }

  const inventoryComponent = myEntity.getComponent("inventory");
  if (!inventoryComponent) {
    return null;
  }

  const handleEquip = (itemId: string) => {
    props.game.handleEquipItem(itemId);
  };

  const handleUnequip = (slot: string) => {
    props.game.handleUnequipSlot(slot);
  };

  const equippedEntries = Object.entries(equippedSlots);

  return (
    <div className={`${panelStyles.panel} ${styles.container}`}>
      <div className={panelStyles.panelHeader}>Inventory</div>
      <div className={panelStyles.panelContent}>
        <div className={styles.sectionHeader}>Equipped</div>
        {equippedEntries.length === 0 ? (
          <div className={styles.emptyMessage}>No equipment</div>
        ) : (
          <div className={styles.equippedList}>
            {equippedEntries.map(([slot, item]) => (
              <div key={slot} className={styles.equippedItem}>
                <div className={styles.equippedSlot}>{slot}</div>
                {item ? (
                  <div className={styles.equippedDetails}>
                    <span className={styles.equippedName}>{item.name}</span>
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
        <div className={styles.sectionHeader}>Backpack</div>
        {items.length === 0 ? (
          <div className={styles.emptyMessage}>Inventory is empty</div>
        ) : (
          <div className={styles.itemsGrid}>
            {items.map((item) => (
              <div
                key={item.id}
                className={`${styles.item} ${
                  item.equipmentSlot ? styles.equipable : ""
                }`}
                title={`${item.name} (${item.type})${
                  item.equipmentSlot ? ` - Equipable: ${item.equipmentSlot}` : ""
                }${
                  item.combatStats
                    ? ` - Dmg ${item.combatStats.minDamage}-${item.combatStats.maxDamage}, Acc ${item.combatStats.accuracyBonus}, Armor ${item.combatStats.armorBonus}`
                    : ""
                }`}
              >
                <div className={styles.itemName}>{item.name}</div>
                <div className={styles.itemType}>{item.type}</div>
                {item.equipmentSlot && (
                  <div className={styles.itemSlot}>{item.equipmentSlot}</div>
                )}
                {item.equipmentSlot && (
                  <button
                    className={styles.equipButton}
                    onClick={() => handleEquip(item.id)}
                  >
                    Equip
                  </button>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
