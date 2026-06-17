import { type SyntheticEvent, useState } from "react";
import Game from "../game/game";
import { ChatBoxContent } from "./components/chatBox";
import InteractionMenu from "./components/interactionMenu";
import { EquipmentContent, InventoryBackpackContent } from "./components/inventory";
import { CombatLogContent } from "./components/combatLog";
import ConversationPanel from "./components/conversationPanel";
import { QuestPanelContent } from "./components/questPanel";
import panelStyles from "./components/uiPanel.module.css";
import styles from "./uiRoot.module.css";

type Props = {
  game: Game;
};

type LeftTab = "chat" | "combat";
type RightTab = "inventory" | "equipment" | "quests";

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
  const stopHudEvent = (event: SyntheticEvent) => {
    event.stopPropagation();
  };
  const handleHudMouseEnter = () => {
    props.game.setPointerOverUi(true);
  };
  const handleHudMouseLeave = () => {
    props.game.setPointerOverUi(false);
  };

  return (
    <div
      className={styles.root}
    >
      <div
        className={`${panelStyles.panel} ${styles.hudPanel} ${styles.leftPanel}`}
        onClick={stopHudEvent}
        onContextMenu={stopHudEvent}
        onMouseDown={stopHudEvent}
        onMouseUp={stopHudEvent}
        onMouseEnter={handleHudMouseEnter}
        onMouseLeave={handleHudMouseLeave}
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
        onMouseDown={stopHudEvent}
        onMouseUp={stopHudEvent}
        onMouseEnter={handleHudMouseEnter}
        onMouseLeave={handleHudMouseLeave}
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
    </div>
  );
}
