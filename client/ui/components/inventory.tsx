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
};

export default function Inventory(props: Props) {
  const [items, setItems] = useState<InventoryItem[]>([]);

  const updateInventory = () => {
    const myEntity = props.game.getMyEntity();
    if (!myEntity) {
      setItems([]);
      return;
    }

    const inventoryComponent = myEntity.getComponent("inventory");
    if (!inventoryComponent || !inventoryComponent.items) {
      setItems([]);
      return;
    }

    setItems(inventoryComponent.items || []);
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

  return (
    <div className={`${panelStyles.panel} ${styles.container}`}>
      <div className={panelStyles.panelHeader}>Inventory</div>
      <div className={panelStyles.panelContent}>
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
                }`}
              >
                <div className={styles.itemName}>{item.name}</div>
                <div className={styles.itemType}>{item.type}</div>
                {item.equipmentSlot && (
                  <div className={styles.itemSlot}>{item.equipmentSlot}</div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

