import * as THREE from "three";
import RendererHuman from "./renderer/rendererHuman";
import RendererChatMessage from "./renderer/rendererChatMessage";
import RendererCombatText from "./renderer/rendererCombatText";
import EntityRenderer from "./renderer/renderer";
import Entity from "./entity/entity";
import RendererBuilding from "./renderer/rendererBuilding";
import RendererChest from "./renderer/rendererChest";
import RendererDoor from "./renderer/rendererDoor";
import RendererError from "./renderer/rendererError";
import RendererRock from "./renderer/rendererRock";
import RendererTree from "./renderer/rendererTree";
import RendererRewardDrop from "./renderer/rendererRewardDrop";
import RendererRat from "./renderer/rendererRat";
import RendererFishingSpot from "./renderer/rendererFishingSpot";
import type { TerrainHeightSampler } from "./renderer/renderer";

type VisualHeightWorld = {
  getVisualHeightAtWorldPosition(worldX: number, worldZ: number): number;
};

type TimedChatEffect = {
  effect: RendererChatMessage;
  timeoutId: number;
};

type TimedCombatEffect = {
  effect: RendererCombatText;
  targetEntityId: string | null;
  timeoutId: number;
};

type PendingChatEffect = {
  message: string;
  expiresAt: number;
};

type PendingCombatEffect = {
  attackerEntityId: string;
  targetEntityId: string;
  didHit: boolean;
  damage: number;
  isCritical: boolean;
  attackPlayed: boolean;
  textShown: boolean;
  expiresAt: number;
};

type RecentCombatAnchor = {
  position: THREE.Vector3;
  expiresAt: number;
};

const CHAT_DISPLAY_MILLISECONDS = 5000;
const COMBAT_TEXT_DISPLAY_MILLISECONDS = 2000;
const EFFECT_RETRY_MILLISECONDS = 1000;
const RECENT_COMBAT_ANCHOR_MILLISECONDS = 1000;
const COMBAT_TEXT_LOCAL_POSITION = new THREE.Vector3(0.5, 1.8, 0.5);

export default class EntityRenderSystem {
  scene: THREE.Scene;
  renderers: Record<string, EntityRenderer | null>;
  private sampleVisualHeight: TerrainHeightSampler;
  private readonly getEstimatedServerTick: () => number;
  private effectsRoot = new THREE.Group();
  private entitiesById = new Map<string, Entity>();
  private chatEffects = new Map<string, TimedChatEffect>();
  private combatEffects = new Set<TimedCombatEffect>();
  private pendingChatEffects = new Map<string, PendingChatEffect>();
  private pendingCombatEffects: PendingCombatEffect[] = [];
  private recentCombatAnchors = new Map<string, RecentCombatAnchor>();

  constructor(
    scene: THREE.Scene,
    getWorld?: () => VisualHeightWorld | undefined,
    getEstimatedServerTick: () => number = () => 0,
  ) {
    this.scene = scene;
    this.renderers = {};
    this.effectsRoot.name = "transient-effects";
    this.scene.add(this.effectsRoot);
    this.sampleVisualHeight = (worldX: number, worldZ: number) =>
      getWorld?.()?.getVisualHeightAtWorldPosition(worldX, worldZ) ?? 0;
    this.getEstimatedServerTick = getEstimatedServerTick;
  }

  createRenderer(renderableType: string, entity: Entity): EntityRenderer | null {
    switch (renderableType) {
      case "human":
        return new RendererHuman(
          this.scene,
          entity,
          this.sampleVisualHeight,
          (entityId) => this.entitiesById.get(entityId),
          this.getEstimatedServerTick,
        );
      case "rat":
        return new RendererRat(this.scene, entity, this.sampleVisualHeight);
      case "tree":
        return new RendererTree(this.scene, entity, this.sampleVisualHeight);
      case "door":
        return new RendererDoor(this.scene, entity, this.sampleVisualHeight);
      case "chest":
        return new RendererChest(this.scene, entity, this.sampleVisualHeight);
      case "rock":
        return new RendererRock(this.scene, entity, this.sampleVisualHeight);
      case "building":
        return new RendererBuilding(this.scene, entity, this.sampleVisualHeight);
      case "rewarddrop":
        return new RendererRewardDrop(this.scene, entity, this.sampleVisualHeight);
      case "fishingSpot":
        return new RendererFishingSpot(this.scene, entity, this.sampleVisualHeight);
    }
    console.error("unknown renderer type:", renderableType);
    return new RendererError(this.scene, entity, this.sampleVisualHeight);
  }

  update(entities: Entity[], deltaSeconds: number) {
    this.entitiesById = new Map(entities.map((entity) => [entity.getId(), entity]));

    for (const entity of entities) {
      const renderableComponent = entity.getComponent("renderable");
      if (!renderableComponent) {
        continue;
      }

      let renderer = this.renderers[entity.getId()];
      if (!renderer) {
        renderer = this.createRenderer(renderableComponent.type, entity);
        if (renderer) {
          this.renderers[entity.getId()] = renderer;
        } else {
          continue;
        }
      }

      renderer!.update(deltaSeconds);
    }

    for (const entityId of Object.keys(this.renderers)) {
      const renderer = this.renderers[entityId];
      if (!renderer) {
        continue;
      }

      if (!entities.find((e) => e.getId() === entityId)) {
        this.rememberCombatAnchor(entityId, renderer);
        this.clearTransientEffectsFor(entityId);
        renderer.onRemove();
        delete this.renderers[entityId];
      }
    }

    this.flushPendingEffects();
    this.expireRecentCombatAnchors();
  }

  getRenderers(): Record<string, EntityRenderer | null> {
    return this.renderers;
  }

  showChatMessage(fromEntityId: string, message: string) {
    if (this.tryShowChatMessage(fromEntityId, message)) {
      this.pendingChatEffects.delete(fromEntityId);
    } else {
      this.pendingChatEffects.set(fromEntityId, {
        message,
        expiresAt: Date.now() + EFFECT_RETRY_MILLISECONDS,
      });
    }
  }

  showCombatResult(
    attackerEntityId: string,
    targetEntityId: string,
    didHit: boolean,
    damage: number,
    isCritical: boolean,
  ) {
    const pending: PendingCombatEffect = {
      attackerEntityId,
      targetEntityId,
      didHit,
      damage,
      isCritical,
      attackPlayed: false,
      textShown: false,
      expiresAt: Date.now() + EFFECT_RETRY_MILLISECONDS,
    };
    if (!this.tryShowCombatResult(pending)) {
      this.pendingCombatEffects.push(pending);
    }
  }

  clearTransientEffects() {
    for (const entityId of [...this.chatEffects.keys()]) {
      this.removeChatEffect(entityId);
    }
    for (const effect of [...this.combatEffects]) {
      this.removeCombatEffect(effect);
    }
    this.pendingChatEffects.clear();
    this.pendingCombatEffects = [];
    this.recentCombatAnchors.clear();
  }

  private clearTransientEffectsFor(entityId: string) {
    this.removeChatEffect(entityId);
    this.detachCombatEffectsFrom(entityId);
    this.pendingChatEffects.delete(entityId);
  }

  private tryShowChatMessage(fromEntityId: string, message: string): boolean {
    const parent = this.renderers[fromEntityId]?.getObject3D();
    if (!parent) {
      return false;
    }

    this.removeChatEffect(fromEntityId);
    const effect = new RendererChatMessage(parent, message);
    const timeoutId = window.setTimeout(() => {
      const current = this.chatEffects.get(fromEntityId);
      if (current?.effect === effect) {
        this.removeChatEffect(fromEntityId);
      }
    }, CHAT_DISPLAY_MILLISECONDS);
    this.chatEffects.set(fromEntityId, { effect, timeoutId });
    return true;
  }

  private tryShowCombatResult(effect: PendingCombatEffect): boolean {
    if (!effect.attackPlayed) {
      const attacker = this.renderers[effect.attackerEntityId];
      if (attacker) {
        attacker.playAttackAnimation();
        effect.attackPlayed = true;
      }
    }

    if (!effect.textShown) {
      const target = this.renderers[effect.targetEntityId]?.getObject3D();
      const recentAnchor = this.recentCombatAnchors.get(effect.targetEntityId);
      if (target || (recentAnchor && recentAnchor.expiresAt > Date.now())) {
        const text = effect.didHit
          ? effect.isCritical
            ? `CRIT ${effect.damage}`
            : `${effect.damage}`
          : "MISS";
        const kind = effect.didHit ? (effect.isCritical ? "crit" : "hit") : "miss";
        const parent = target ?? this.effectsRoot;
        const position = target
          ? COMBAT_TEXT_LOCAL_POSITION.clone()
          : this.effectsRoot.worldToLocal(recentAnchor!.position.clone());
        const renderer = new RendererCombatText(parent, position, text, kind);
        const timedEffect: TimedCombatEffect = {
          effect: renderer,
          targetEntityId: target ? effect.targetEntityId : null,
          timeoutId: 0,
        };
        timedEffect.timeoutId = window.setTimeout(() => {
          this.removeCombatEffect(timedEffect);
        }, COMBAT_TEXT_DISPLAY_MILLISECONDS);
        this.combatEffects.add(timedEffect);
        effect.textShown = true;
      }
    }

    return effect.attackPlayed && effect.textShown;
  }

  private flushPendingEffects() {
    const now = Date.now();
    for (const [entityId, effect] of this.pendingChatEffects) {
      if (effect.expiresAt <= now || this.tryShowChatMessage(entityId, effect.message)) {
        this.pendingChatEffects.delete(entityId);
      }
    }

    this.pendingCombatEffects = this.pendingCombatEffects.filter(
      (effect) => effect.expiresAt > now && !this.tryShowCombatResult(effect),
    );
  }

  private rememberCombatAnchor(entityId: string, renderer: EntityRenderer) {
    const object = renderer.getObject3D();
    if (!object) {
      return;
    }
    this.recentCombatAnchors.set(entityId, {
      position: object.localToWorld(COMBAT_TEXT_LOCAL_POSITION.clone()),
      expiresAt: Date.now() + RECENT_COMBAT_ANCHOR_MILLISECONDS,
    });
  }

  private expireRecentCombatAnchors() {
    const now = Date.now();
    for (const [entityId, anchor] of this.recentCombatAnchors) {
      if (anchor.expiresAt <= now) {
        this.recentCombatAnchors.delete(entityId);
      }
    }
  }

  private detachCombatEffectsFrom(entityId: string) {
    for (const effect of this.combatEffects) {
      if (effect.targetEntityId !== entityId) {
        continue;
      }
      effect.effect.reparent(this.effectsRoot);
      effect.targetEntityId = null;
    }
  }

  private removeChatEffect(entityId: string) {
    const current = this.chatEffects.get(entityId);
    if (!current) {
      return;
    }
    window.clearTimeout(current.timeoutId);
    current.effect.dispose();
    this.chatEffects.delete(entityId);
  }

  private removeCombatEffect(effect: TimedCombatEffect) {
    window.clearTimeout(effect.timeoutId);
    effect.effect.dispose();
    this.combatEffects.delete(effect);
  }
}
