import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import Game from "../../game/game";
import { BankUpdateEventName } from "../../events/bank";
import { getItemIconSrc, getItemTitle, type InventoryItem } from "./inventory";
import inventoryStyles from "./inventory.module.css";
import panelStyles from "./uiPanel.module.css";
import styles from "./bankPanel.module.css";

type BankContainer = "bank" | "backpack";
type BankState = { targetId: string; bank: InventoryItem[]; backpack: InventoryItem[] };
type TransferMenu = { container: BankContainer; itemId: string; x: number; y: number; custom: boolean };

function readBank(game: Game): BankState | null {
  const player = game.getMyEntity();
  const targetId = game.getBankPanelTargetId();
  const bank = player?.getComponent("bank");
  if (!targetId || !bank) return null;
  return { targetId, bank: bank.items,
    backpack: player?.getComponent("inventory")?.items ?? [] };
}

export default function BankPanel({ game }: { game: Game }) {
  const [bank, setBank] = useState(() => readBank(game));
  const [menu, setMenu] = useState<TransferMenu | null>(null);
  const [quantity, setQuantity] = useState("1");
  const closeRef = useRef<HTMLButtonElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement | null>(null);
  const targetId = bank?.targetId;
  const selectedItem = menu ? bank?.[menu.container].find(item => item.id === menu.itemId) : undefined;

  const closeMenu = () => {
    setMenu(null);
    if (triggerRef.current?.isConnected) triggerRef.current.focus({ preventScroll: true });
  };

  const openMenu = (container: BankContainer, item: InventoryItem, button: HTMLButtonElement, x?: number, y?: number) => {
    const bounds = button.getBoundingClientRect();
    triggerRef.current = button;
    setQuantity("1");
    setMenu({ container, itemId: item.id, x: x ?? bounds.left, y: y ?? bounds.bottom, custom: false });
  };

  useEffect(() => {
    const update = () => setBank(readBank(game));
    game.addEventListener(BankUpdateEventName, update);
    return () => game.removeEventListener(BankUpdateEventName, update);
  }, [game]);

  useEffect(() => {
    setMenu(null);
    if (!targetId) return;
    const previousFocus = document.activeElement;
    closeRef.current?.focus({ preventScroll: true });
    const keydown = (event: KeyboardEvent) => {
      if (event.key !== "Escape") return;
      event.preventDefault();
      if (menuRef.current) closeMenu();
      else game.handleBankClose(targetId);
    };
    window.addEventListener("keydown", keydown);
    return () => {
      window.removeEventListener("keydown", keydown);
      game.setPointerOverUi(false);
      if (previousFocus instanceof HTMLElement && previousFocus.isConnected) previousFocus.focus({ preventScroll: true });
    };
  }, [game, targetId]);

  useLayoutEffect(() => {
    const element = menuRef.current;
    if (!menu || !element) return;
    const bounds = element.getBoundingClientRect();
    element.style.left = `${Math.max(8, Math.min(menu.x, window.innerWidth - bounds.width - 8))}px`;
    element.style.top = `${Math.max(8, Math.min(menu.y, window.innerHeight - bounds.height - 8))}px`;
    const input = element.querySelector<HTMLInputElement>("input");
    (input ?? element.querySelector<HTMLButtonElement>("button"))?.focus({ preventScroll: true });
    input?.select();
  }, [menu]);

  useEffect(() => {
    if (menu && !selectedItem) setMenu(null);
  }, [menu, selectedItem]);

  useEffect(() => {
    if (!menu) return;
    const dismiss = () => setMenu(null);
    const pointerdown = (event: PointerEvent) => {
      if (!menuRef.current?.contains(event.target as Node)) dismiss();
    };
    window.addEventListener("pointerdown", pointerdown, true);
    window.addEventListener("resize", dismiss);
    window.addEventListener("scroll", dismiss, true);
    return () => {
      window.removeEventListener("pointerdown", pointerdown, true);
      window.removeEventListener("resize", dismiss);
      window.removeEventListener("scroll", dismiss, true);
    };
  }, [menu]);

  if (!bank) return null;
  const transfer = (container: BankContainer, item: InventoryItem, count: number) => {
    if (menu) closeMenu();
    game.handleBankTransfer(bank.targetId, container === "bank" ? "withdraw" : "deposit", item.id, count);
  };
  const menuAction = menu?.container === "bank" ? "Withdraw" : "Deposit";
  const count = Number(quantity);
  const validQuantity = Number.isSafeInteger(count) && count > 0 && count <= (selectedItem?.quantity ?? 0);

  return (
    <section className={`${panelStyles.panel} ${styles.container}`} role="dialog" aria-labelledby="bank-title"
      onClick={(event) => event.stopPropagation()}
      onContextMenu={(event) => { event.preventDefault(); event.stopPropagation(); }}
      onPointerDown={(event) => { event.stopPropagation(); game.setPointerOverUi(true); }}
      onPointerMove={(event) => event.stopPropagation()}
      onPointerEnter={() => game.setPointerOverUi(true)}
      onPointerLeave={() => game.setPointerOverUi(false)}
      onPointerUp={(event) => { if (event.pointerType !== "mouse") game.setPointerOverUi(false); }}>
      <div className={`${panelStyles.panelHeader} ${styles.header}`}>
        <span id="bank-title">Bank</span>
        <button ref={closeRef} type="button" className={styles.close} aria-label="Close bank"
          onClick={() => game.handleBankClose(bank.targetId)}>×</button>
      </div>
      <div className={styles.columns}>
        {(["bank", "backpack"] as const).map(container => {
          const label = container === "bank" ? "Withdraw" : "Deposit";
          const items = bank[container];
          const slotCount = container === "bank" ? Math.max(30, Math.ceil((items.length + 1) / 6) * 6) : 20;
          return <section key={container} className={styles.column} aria-label={container === "bank" ? "Bank items" : "Backpack items"}>
            <h2>{container === "bank" ? "Bank" : "Inventory"}<span>{container === "bank" ? items.length : `${items.length}/20`}</span></h2>
            <div className={`${styles.grid} ${container === "bank" ? styles.bankGrid : styles.backpackGrid}`}>
              {Array.from({ length: slotCount }, (_, index) => {
                const item = items[index];
                return item ? <button key={item.id} type="button"
                  className={`${inventoryStyles.item} ${styles.slot} ${item.equipmentSlot ? inventoryStyles.equipable : ""}`}
                  title={getItemTitle(item)} aria-label={`${label} ${item.name}, quantity ${item.quantity}`}
                  aria-haspopup="menu" aria-expanded={menu?.container === container && menu.itemId === item.id}
                  onContextMenu={(event) => {
                    event.preventDefault(); event.stopPropagation();
                    openMenu(container, item, event.currentTarget, event.clientX || undefined, event.clientY || undefined);
                  }}
                  onKeyDown={(event) => {
                    if (event.key === "ContextMenu" || (event.shiftKey && event.key === "F10")) {
                      event.preventDefault(); openMenu(container, item, event.currentTarget);
                    }
                  }}
                  onClick={(event) => {
                    if (event.detail === 0 || (event.nativeEvent instanceof PointerEvent && event.nativeEvent.pointerType === "touch")) {
                      openMenu(container, item, event.currentTarget);
                    } else transfer(container, item, 1);
                  }}>
                  <img className={styles.icon} src={getItemIconSrc(item)} alt="" draggable={false} />
                  {item.stackable && <span className={inventoryStyles.quantity}>{item.quantity.toLocaleString()}</span>}
                </button> : <div key={`empty-${index}`} className={`${inventoryStyles.item} ${inventoryStyles.emptySlot} ${styles.slot}`}
                  aria-label={`Empty ${container === "bank" ? "bank" : "inventory"} slot ${index + 1}`} />;
              })}
            </div>
          </section>;
        })}
      </div>
      {menu && selectedItem && createPortal(
        <div ref={menuRef} className={`${inventoryStyles.itemMenu} ${styles.menu}`}
          role={menu.custom ? "dialog" : "menu"} aria-label={`${menuAction} ${selectedItem.name}`}
          style={{ left: menu.x, top: menu.y }}
          onContextMenu={event => { event.preventDefault(); event.stopPropagation(); }}
          onClick={event => event.stopPropagation()}
          onPointerDown={event => { event.stopPropagation(); game.setPointerOverUi(true); }}
          onPointerEnter={() => game.setPointerOverUi(true)}
          onPointerLeave={() => game.setPointerOverUi(false)}
          onKeyDown={event => {
            if (event.key === "Escape") { event.preventDefault(); event.stopPropagation(); closeMenu(); return; }
            if (event.key === "Tab" && !menu.custom) { closeMenu(); return; }
            if (menu.custom || !["ArrowDown", "ArrowUp", "Home", "End"].includes(event.key)) return;
            event.preventDefault(); event.stopPropagation();
            const buttons = Array.from(event.currentTarget.querySelectorAll<HTMLButtonElement>("button"));
            const current = buttons.indexOf(document.activeElement as HTMLButtonElement);
            const next = event.key === "Home" ? 0 : event.key === "End" ? buttons.length - 1 :
              (current + (event.key === "ArrowDown" ? 1 : -1) + buttons.length) % buttons.length;
            buttons[next]?.focus();
          }}>
          <div className={inventoryStyles.itemMenuName}>{selectedItem.name}</div>
          {menu.custom ? <form className={styles.quantityForm} onSubmit={event => {
            event.preventDefault();
            if (validQuantity) transfer(menu.container, selectedItem, count);
          }}>
            <label htmlFor="bank-quantity">Quantity</label>
            <input id="bank-quantity" type="number" min="1" max={selectedItem.quantity} step="1" value={quantity}
              onChange={event => setQuantity(event.target.value)} />
            <div className={styles.quantityActions}>
              <button type="submit" disabled={!validQuantity}>{menuAction}</button>
              <button type="button" onClick={closeMenu}>Cancel</button>
            </div>
          </form> : <>
            <button role="menuitem" onClick={() => transfer(menu.container, selectedItem, 1)}>{menuAction} 1</button>
            <button role="menuitem" onClick={() => setMenu({ ...menu, custom: true })}>{menuAction} X</button>
            <button role="menuitem" onClick={() => transfer(menu.container, selectedItem, selectedItem.quantity)}>{menuAction} All</button>
          </>}
        </div>, document.getElementById("uiLayerRoot")!,
      )}
    </section>
  );
}
