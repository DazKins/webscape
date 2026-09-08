import { useEffect, useRef, useState } from "react";
import Game from "../../game/game";
import { ShopUpdateEventName, TradeResultEventName, type TradeResultEvent } from "../../events/shop";
import { getItemIconSrc, type InventoryItem } from "./inventory";
import panelStyles from "./uiPanel.module.css";
import styles from "./shopPanel.module.css";

type Offer = { itemId: string; buyPrice: number; sellPrice: number; item: InventoryItem };
type ShopState = { targetId: string; name: string; offers: Offer[]; items: InventoryItem[] };

function readShop(game: Game): ShopState | null {
  const player = game.getMyEntity();
  const targetId = player?.getComponent("trading")?.targetEntityId;
  const shop = targetId ? game.getEntity(targetId)?.getComponent("shop") : null;
  if (!shop) return null;
  return { targetId, name: game.getEntityName(targetId, "Shop"), offers: shop.offers,
    items: player?.getComponent("inventory")?.items ?? [] };
}

function matchesOffer(item: InventoryItem, offer: Offer): boolean {
  const other = offer.item;
  return item.name === other.name && item.type === other.type && item.renderModel === other.renderModel &&
    item.equipmentSlot === other.equipmentSlot && JSON.stringify(item.combatStats) === JSON.stringify(other.combatStats);
}

export default function ShopPanel({ game }: { game: Game }) {
  const [shop, setShop] = useState(() => readShop(game));
  const [tab, setTab] = useState<"buy" | "sell">("buy");
  const [pending, setPending] = useState(false);
  const [feedback, setFeedback] = useState("");
  const closeRef = useRef<HTMLButtonElement>(null);
  const itemsRef = useRef<HTMLDivElement>(null);
  const targetId = shop?.targetId;

  useEffect(() => {
    if (itemsRef.current) itemsRef.current.scrollTop = 0;
  }, [tab, targetId]);

  useEffect(() => {
    const update = () => setShop(readShop(game));
    game.addEventListener(ShopUpdateEventName, update);
    return () => game.removeEventListener(ShopUpdateEventName, update);
  }, [game]);

  useEffect(() => {
    setPending(false);
    setFeedback("");
    setTab("buy");
    if (!targetId) { game.setPointerOverUi(false); return; }
    const previousFocus = document.activeElement;
    closeRef.current?.focus({ preventScroll: true });
    const keydown = (event: KeyboardEvent) => {
      if (event.key === "Escape") { event.preventDefault(); game.handleTradeClose(targetId); }
    };
    const result = (event: Event) => {
      const payload = (event as TradeResultEvent).payload;
      if (payload.targetEntityId !== targetId) return;
      setPending(false);
      setFeedback(payload.text);
    };
    window.addEventListener("keydown", keydown);
    game.addEventListener(TradeResultEventName, result);
    return () => {
      window.removeEventListener("keydown", keydown);
      game.removeEventListener(TradeResultEventName, result);
      game.setPointerOverUi(false);
      if (previousFocus instanceof HTMLElement && previousFocus.isConnected) previousFocus.focus({ preventScroll: true });
    };
  }, [game, targetId]);

  if (!shop) return null;
  const gold = shop.items.find((item) => item.type === "gold")?.quantity ?? 0;
  const sellable = shop.items.flatMap((item) => {
    const offer = shop.offers.find((candidate) => matchesOffer(item, candidate));
    return offer ? [{ item, offer }] : [];
  });
  const trade = (action: "buy" | "sell", id: string) => {
    setPending(true);
    setFeedback("");
    game.handleTrade(shop.targetId, action, id);
  };
  const rows = tab === "buy"
    ? shop.offers.map((offer) => ({ item: offer.item, offer })) : sellable;

  return (
    <section className={`${panelStyles.panel} ${styles.container}`} role="dialog" aria-labelledby="shop-title"
      onClick={(event) => event.stopPropagation()}
      onContextMenu={(event) => event.stopPropagation()}
      onPointerDown={(event) => { event.stopPropagation(); game.setPointerOverUi(true); }}
      onPointerMove={(event) => event.stopPropagation()}
      onPointerEnter={() => game.setPointerOverUi(true)}
      onPointerLeave={() => game.setPointerOverUi(false)}
      onPointerUp={(event) => { if (event.pointerType !== "mouse") game.setPointerOverUi(false); }}>
      <div className={`${panelStyles.panelHeader} ${styles.header}`}>
        <span id="shop-title">{shop.name}</span>
        <button ref={closeRef} type="button" className={styles.close} aria-label="Close shop"
          onClick={() => game.handleTradeClose(shop.targetId)}>×</button>
      </div>
      <div className={styles.balance}><span>Your gold</span><strong>{gold.toLocaleString()} <span aria-hidden="true">●</span></strong></div>
      <div className={styles.tabs} aria-label="Shop actions">
        <button type="button" aria-pressed={tab === "buy"} onClick={() => { setTab("buy"); setFeedback(""); }}>Buy goods</button>
        <button type="button" aria-pressed={tab === "sell"} onClick={() => { setTab("sell"); setFeedback(""); }}>Sell from backpack</button>
      </div>
      {tab === "sell" && <p className={styles.hint}>Sell gathered materials and spare gear for gold.</p>}
      <div ref={itemsRef} className={styles.items}>
        {rows.map(({ item, offer }) => {
          const price = tab === "buy" ? offer.buyPrice : offer.sellPrice;
          const full = tab === "buy" && shop.items.length >= 20 && gold !== price;
          const disabled = pending || (tab === "buy" && (gold < price || full));
          return <div className={styles.row} key={tab === "buy" ? offer.itemId : item.id}>
            <img src={getItemIconSrc(item)} alt="" draggable={false} />
            <div className={styles.itemText}><strong>{item.name}</strong><span>{price} gold</span></div>
            <button type="button" disabled={disabled} aria-label={`${tab === "buy" ? "Buy" : "Sell"} ${item.name} for ${price} gold`}
              title={full ? "Your backpack is full" : tab === "buy" && gold < price ? "Not enough gold" : undefined}
              onClick={() => trade(tab, tab === "buy" ? offer.itemId : item.id)}>{tab === "buy" ? "Buy" : "Sell"}</button>
          </div>;
        })}
        {rows.length === 0 && <p className={styles.empty}>No items this shop buys. Try gathering logs or catching fish.</p>}
      </div>
      <div className={styles.footer} role="status" aria-live="polite">{pending ? "Trading…" : feedback || `${shop.items.length}/20 backpack slots`}</div>
    </section>
  );
}
