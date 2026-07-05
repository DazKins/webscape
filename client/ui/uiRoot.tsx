import { type SyntheticEvent, useEffect, useState } from "react";
import Game from "../game/game";
import { ChatBoxContent } from "./components/chatBox";
import InteractionMenu from "./components/interactionMenu";
import { EquipmentContent, InventoryBackpackContent } from "./components/inventory";
import { CombatLogContent } from "./components/combatLog";
import ConversationPanel from "./components/conversationPanel";
import { QuestPanelContent } from "./components/questPanel";
import QuestCompletedOverlay from "./components/questCompletedOverlay";
import panelStyles from "./components/uiPanel.module.css";
import styles from "./uiRoot.module.css";
import { getDeviceProfile, type DeviceProfile } from "../responsive";

type Props = {
  game: Game;
};

type LeftTab = "chat" | "combat";
type RightTab = "inventory" | "equipment" | "quests";
type MobileTab = LeftTab | RightTab;
type SheetState = "collapsed" | "half" | "expanded";

type TabDefinition<T extends string> = {
  id: T;
  label: string;
  icon: string;
};

function createIcon(body: string, background: string, foreground: string): string {
  const svg = `
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48">
      <rect width="48" height="48" rx="10" fill="${background}"/>
      ${body.split("currentColor").join(foreground)}
    </svg>
  `;
  return `data:image/svg+xml,${encodeURIComponent(svg)}`;
}

const tabIcons = {
  chat: createIcon(
    `<path d="M12 14h24v15H23l-7 6v-6h-4V14Z" fill="none" stroke="currentColor" stroke-width="4" stroke-linejoin="round"/>`,
    "#1f3439",
    "#9ee7f2",
  ),
  combat: createIcon(
    `<path d="M32 9 39 16 18 37l-7 2 2-7L32 9Z" fill="currentColor"/><path d="M13 13 35 35" stroke="currentColor" stroke-width="4" stroke-linecap="round"/>`,
    "#3a2325",
    "#ffaaa0",
  ),
  inventory: createIcon(
    `<path d="M14 17h20l4 6v16H10V23l4-6Z" fill="none" stroke="currentColor" stroke-width="4" stroke-linejoin="round"/><path d="M18 17c0-5 3-8 6-8s6 3 6 8" fill="none" stroke="currentColor" stroke-width="4" stroke-linecap="round"/>`,
    "#293244",
    "#b7ccff",
  ),
  equipment: createIcon(
    `<path d="M24 8 37 13v11c0 9-5 15-13 19-8-4-13-10-13-19V13l13-5Z" fill="none" stroke="currentColor" stroke-width="4" stroke-linejoin="round"/><path d="M24 13v25" stroke="currentColor" stroke-width="4" stroke-linecap="round"/>`,
    "#253831",
    "#a9e6b2",
  ),
  quests: createIcon(
    `<path d="M15 10h18c4 0 6 3 4 7v22H17c-5 0-8-4-5-9V10Z" fill="none" stroke="currentColor" stroke-width="4" stroke-linejoin="round"/><path d="M17 25h15M17 32h12" stroke="currentColor" stroke-width="4" stroke-linecap="round"/>`,
    "#3a3023",
    "#f0cf91",
  ),
};

const leftTabs: TabDefinition<LeftTab>[] = [
  { id: "chat", label: "Chat", icon: tabIcons.chat },
  { id: "combat", label: "Combat Log", icon: tabIcons.combat },
];

const rightTabs: TabDefinition<RightTab>[] = [
  { id: "inventory", label: "Inventory", icon: tabIcons.inventory },
  { id: "equipment", label: "Equipment", icon: tabIcons.equipment },
  { id: "quests", label: "Quests", icon: tabIcons.quests },
];

const mobileTabs: TabDefinition<MobileTab>[] = [...leftTabs, ...rightTabs];

function getWindowProfile() {
  return getDeviceProfile({
    width: Math.max(1, window.innerWidth),
    height: Math.max(1, window.innerHeight),
  });
}

function useDeviceProfile(game: Game): DeviceProfile {
  const [profile, setProfile] = useState<DeviceProfile>(() => game.getDeviceProfile());

  useEffect(() => {
    const update = () => setProfile(getWindowProfile());
    const pointerQuery = window.matchMedia("(pointer: coarse)");
    const hoverQuery = window.matchMedia("(hover: hover)");

    update();
    window.addEventListener("resize", update);
    window.addEventListener("orientationchange", update);
    pointerQuery.addEventListener("change", update);
    hoverQuery.addEventListener("change", update);

    return () => {
      window.removeEventListener("resize", update);
      window.removeEventListener("orientationchange", update);
      pointerQuery.removeEventListener("change", update);
      hoverQuery.removeEventListener("change", update);
    };
  }, [game]);

  return profile;
}

function TabButtons<T extends string>(props: {
  tabs: TabDefinition<T>[];
  activeTab: T;
  setActiveTab: (tab: T) => void;
  label: string;
}) {
  return (
    <div className={styles.tabs} role="tablist" aria-label={props.label}>
      {props.tabs.map((tab) => (
        <button
          key={tab.id}
          className={`${styles.tabButton} ${
            props.activeTab === tab.id ? styles.activeTab : ""
          }`}
          type="button"
          role="tab"
          aria-label={tab.label}
          aria-selected={props.activeTab === tab.id}
          title={tab.label}
          onClick={() => props.setActiveTab(tab.id)}
        >
          <img className={styles.tabIcon} src={tab.icon} alt="" draggable={false} />
        </button>
      ))}
    </div>
  );
}

export default function UiRoot(props: Props) {
  const [leftTab, setLeftTab] = useState<LeftTab>("chat");
  const [rightTab, setRightTab] = useState<RightTab>("inventory");
  const [mobileTab, setMobileTab] = useState<MobileTab>("chat");
  const [sheetState, setSheetState] = useState<SheetState>("collapsed");
  const profile = useDeviceProfile(props.game);

  const stopHudEvent = (event: SyntheticEvent) => {
    event.stopPropagation();
  };
  const handleHudMouseEnter = () => {
    props.game.setPointerOverUi(true);
  };
  const handleHudMouseLeave = () => {
    props.game.setPointerOverUi(false);
  };
  const handleHudPointerDown = (event: SyntheticEvent) => {
    event.stopPropagation();
    props.game.setPointerOverUi(true);
  };
  const handleHudPointerUp = (event: SyntheticEvent) => {
    event.stopPropagation();
    const pointerEvent = event.nativeEvent as PointerEvent;
    if (pointerEvent.pointerType !== "mouse") {
      window.setTimeout(() => props.game.setPointerOverUi(false), 0);
    }
  };

  const handleMobileTab = (tab: MobileTab) => {
    setMobileTab(tab);
    setSheetState((current) => (current === "collapsed" ? "half" : current));
  };

  const renderMobileContent = () => {
    switch (mobileTab) {
      case "chat":
        return <ChatBoxContent game={props.game} />;
      case "combat":
        return <CombatLogContent game={props.game} />;
      case "inventory":
        return <InventoryBackpackContent game={props.game} />;
      case "equipment":
        return <EquipmentContent game={props.game} />;
      case "quests":
        return <QuestPanelContent game={props.game} />;
    }
  };

  const activeMobileLabel =
    mobileTabs.find((tab) => tab.id === mobileTab)?.label ?? "Menu";

  if (profile.isMobileLayout) {
    return (
      <div className={styles.root}>
        <div
          className={`${panelStyles.panel} ${styles.mobileSheet} ${styles[sheetState]}`}
          onClick={stopHudEvent}
          onContextMenu={stopHudEvent}
          onPointerDown={handleHudPointerDown}
          onPointerUp={handleHudPointerUp}
          onPointerEnter={handleHudMouseEnter}
          onPointerLeave={handleHudMouseLeave}
        >
          <div className={styles.mobileSheetHeader}>
            <button
              className={styles.sheetToggle}
              type="button"
              aria-label={sheetState === "collapsed" ? "Open panel" : "Collapse panel"}
              onClick={() =>
                setSheetState((current) => (current === "collapsed" ? "half" : "collapsed"))
              }
            >
              {activeMobileLabel}
            </button>
            <button
              className={styles.sheetSizeButton}
              type="button"
              aria-label={sheetState === "expanded" ? "Reduce panel" : "Expand panel"}
              onClick={() =>
                setSheetState((current) => (current === "expanded" ? "half" : "expanded"))
              }
            >
              {sheetState === "expanded" ? "Half" : "Full"}
            </button>
          </div>
          <TabButtons
            tabs={mobileTabs}
            activeTab={mobileTab}
            setActiveTab={handleMobileTab}
            label="Game panels"
          />
          <div className={styles.panelBody}>{renderMobileContent()}</div>
        </div>

        <InteractionMenu game={props.game} />
        <ConversationPanel game={props.game} />
        <QuestCompletedOverlay game={props.game} />
      </div>
    );
  }

  return (
    <div
      className={styles.root}
    >
      <div
        className={`${panelStyles.panel} ${styles.hudPanel} ${styles.leftPanel}`}
        onClick={stopHudEvent}
        onContextMenu={stopHudEvent}
        onPointerDown={handleHudPointerDown}
        onPointerUp={handleHudPointerUp}
        onPointerEnter={handleHudMouseEnter}
        onPointerLeave={handleHudMouseLeave}
      >
        <TabButtons
          tabs={leftTabs}
          activeTab={leftTab}
          setActiveTab={setLeftTab}
          label="Message panels"
        />
        <div className={styles.panelBody}>
          {leftTab === "chat" ? (
            <ChatBoxContent game={props.game} />
          ) : (
            <CombatLogContent game={props.game} />
          )}
        </div>
      </div>

      <div
        className={`${panelStyles.panel} ${styles.hudPanel} ${styles.rightPanel}`}
        onClick={stopHudEvent}
        onContextMenu={stopHudEvent}
        onPointerDown={handleHudPointerDown}
        onPointerUp={handleHudPointerUp}
        onPointerEnter={handleHudMouseEnter}
        onPointerLeave={handleHudMouseLeave}
      >
        <TabButtons
          tabs={rightTabs}
          activeTab={rightTab}
          setActiveTab={setRightTab}
          label="Character panels"
        />
        <div className={styles.panelBody}>
          {rightTab === "inventory" ? (
            <InventoryBackpackContent game={props.game} />
          ) : rightTab === "equipment" ? (
            <EquipmentContent game={props.game} />
          ) : (
            <QuestPanelContent game={props.game} />
          )}
        </div>
      </div>

      <InteractionMenu game={props.game} />
      <ConversationPanel game={props.game} />
      <QuestCompletedOverlay game={props.game} />
    </div>
  );
}
