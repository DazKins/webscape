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
import RendererMagicBolt from "./renderer/rendererMagicBolt";
import RendererArrow from "./renderer/rendererArrow";

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

export type CombatProjectileLaunchedPayload = {
  attackerEntityId: string;
  targetEntityId: string;
  projectileType: string;
  origin: { x: number; y: number };
  targetPosition: { x: number; y: number };
  launchTick: number;
  impactTick: number;
};

type ActiveCombatProjectile = {
  renderer: RendererMagicBolt | RendererArrow;
  targetEntityId: string;
  fallbackTarget: THREE.Vector3;
  launchTick: number;
  impactTick: number;
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
  private readonly getTickSeconds: () => number;
  private effectsRoot = new THREE.Group();
  private entitiesById = new Map<string, Entity>();
  private chatEffects = new Map<string, TimedChatEffect>();
  private combatEffects = new Set<TimedCombatEffect>();
  private pendingChatEffects = new Map<string, PendingChatEffect>();
  private pendingCombatEffects: PendingCombatEffect[] = [];
  private recentCombatAnchors = new Map<string, RecentCombatAnchor>();
  private combatProjectiles = new Set<ActiveCombatProjectile>();

  constructor(
    scene: THREE.Scene,
    getWorld?: () => VisualHeightWorld | undefined,
    getEstimatedServerTick: () => number = () => 0,
    getTickSeconds: () => number = () => 0.5,
  ) {
    this.scene = scene;
    this.renderers = {};
    this.effectsRoot.name = "transient-effects";
    this.scene.add(this.effectsRoot);
    this.sampleVisualHeight = (worldX: number, worldZ: number) =>
      getWorld?.()?.getVisualHeightAtWorldPosition(worldX, worldZ) ?? 0;
    this.getEstimatedServerTick = getEstimatedServerTick;
    this.getTickSeconds = getTickSeconds;
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
          this.getTickSeconds,
        );
      case "rat":
        return new RendererRat(this.scene, entity, this.sampleVisualHeight, this.getTickSeconds);
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
    this.updateCombatProjectiles();
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
    attackMethod: string = "melee",
  ) {
    const pending: PendingCombatEffect = {
      attackerEntityId,
      targetEntityId,
      didHit,
      damage,
      isCritical,
      attackPlayed: attackMethod === "magic" || attackMethod === "ranged",
      textShown: false,
      expiresAt: Date.now() + EFFECT_RETRY_MILLISECONDS,
    };
    if (!this.tryShowCombatResult(pending)) {
      this.pendingCombatEffects.push(pending);
    }
  }

  showCombatProjectile(payload: CombatProjectileLaunchedPayload) {
    if (payload.impactTick <= payload.launchTick) {
      return;
    }
    const authoritativeOrigin = this.worldPosition(payload.origin, 1.1);
    const start = this.renderers[payload.attackerEntityId]
      ?.getProjectileOrigin(payload.projectileType) ?? authoritativeOrigin;
    const renderer = payload.projectileType === "magicBolt"
      ? new RendererMagicBolt(this.effectsRoot, start)
      : payload.projectileType === "arrow"
        ? new RendererArrow(this.effectsRoot, start)
        : null;
    if (!renderer) {
      return;
    }
    const projectile: ActiveCombatProjectile = {
      renderer,
      targetEntityId: payload.targetEntityId,
      fallbackTarget: this.worldPosition(payload.targetPosition, 1.0),
      launchTick: payload.launchTick,
      impactTick: payload.impactTick,
    };
    this.combatProjectiles.add(projectile);
    this.updateCombatProjectile(projectile);
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
    for (const projectile of this.combatProjectiles) {
      projectile.renderer.dispose();
    }
    this.combatProjectiles.clear();
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

  private updateCombatProjectiles() {
    for (const projectile of [...this.combatProjectiles]) {
      this.updateCombatProjectile(projectile);
    }
  }

  private updateCombatProjectile(projectile: ActiveCombatProjectile) {
    const tick = this.getEstimatedServerTick();
    if (tick >= projectile.impactTick) {
      projectile.renderer.dispose();
      this.combatProjectiles.delete(projectile);
      return;
    }
    const duration = projectile.impactTick - projectile.launchTick;
    const progress = THREE.MathUtils.clamp((tick - projectile.launchTick) / duration, 0, 1);
    const targetObject = this.renderers[projectile.targetEntityId]?.getObject3D();
    const target = targetObject
      ? targetObject.localToWorld(new THREE.Vector3(0.5, 1.0, 0.5))
      : projectile.fallbackTarget;
    projectile.renderer.update(progress, target);
  }

  private worldPosition(position: { x: number; y: number }, heightOffset: number) {
    return new THREE.Vector3(
      position.x + 0.5,
      this.sampleVisualHeight(position.x + 0.5, position.y + 0.5) + heightOffset,
      position.y + 0.5,
    );
  }
}
