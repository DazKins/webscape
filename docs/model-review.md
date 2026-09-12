# Model and animation review

The September 2026 pass covers all 32 factories in `client/game/models/registry.ts`, all 21 current animation clips, the wall renderer, and both projectile renderers. Models retain their existing identifiers, tile scale, low-poly style, appearance palettes, and equipment sockets.

| Model | Intended appearance and review result |
| --- | --- |
| human | Tunic-clad adventurer. Added a belt buckle, tunic hem, nose, ears, boot cuffs, and ankle articulation. Retained all four hairstyles and appearance palettes. |
| rat | Low, long-bodied rodent. Connected the legs with furred haunches, corrected the tail taper, lowered and curved its attachment, and softened idle/run bobbing. |
| tree | Harvestable conifer. Added three irregular foliage tiers, roots, and stump growth rings; retained all five damage stages, chips, and chop marks. |
| rock | Faceted boulder. Flattened and grounded its base; introduced slight asymmetry. |
| building | Small timber-framed plaster building. Roof follows both footprint dimensions; added foundation, corner posts, beams, side windows, and door hardware. |
| door | Hinged wooden door. Replaced the buried inset with visible planks, straps, rivets, and handles on both faces. Preserved the existing hinge/orientation convention. |
| chest | Wooden treasure chest. Added a barrel lid, iron bands, rivets, lock, and a real interior. Lid animation follows replicated `openable.isOpen`, including initial load and closing. Chests without that component remain closed. |
| rewarddrop | Tied loot sack. Connected the bag neck and drawstring; replaced the floating square glint with a coin emblem. |
| fishingSpot | Bobber and water ripples. Added restrained bobbing and ripple motion; corrected the ellipse's scale axis. |
| ironSword | Straight iron blade with brass hilt. Replaced the spear-like tip with a continuous tapered blade and point. |
| woodcuttingAxe | Single-bit axe. Widened the blade toward its cutting edge and tapered its thickness into the socket. |
| fishingRod | Wooden rod with reel. Connected the shaft to its tip and added line guides, line along the shaft, and a crank. |
| woodenBow | Simple wooden bow. Balanced and tapered the limbs around the grip; reduced string thickness and added synchronized draw/release. |
| magicStaff | Knotted wooden staff with luminous crystal. Retained its readable silhouette and crown socket. |
| leatherHelmet | Leather cap with band and guards. Retained the existing silhouette and headwear attachment. |
| gold | Coin stacks. Added stamped faces and raised rims to exposed top coins. |
| healthPotion | Corked red potion bottle. Adjusted the bottle's material to read as glass; retained the label. |
| bread | Scored loaf. Retained the readable shape and crust details. |
| apple | Red apple with stem and leaf. Flattened the crown slightly and seated the stem. |
| ironOre | Rock containing metal. Lowered the metal inclusions into the rock rather than leaving tall crystal-like protrusions. |
| stone | Small loose stone. Retained its distinct flattened silhouette. |
| wood | Sawn planks. Retained grain strips and proportions. |
| logs | Stack of cut logs. Retained bark, end grain, and growth rings. |
| arrow | Wooden arrow with iron point. Added a flat tapered point and three radial vanes. The projectile now uses this same model. |
| fish | Caught fish lying on its side. Reversed and flattened the tail so it widens away from the body. |
| mysteriousKey | Brass warded key. Retained the readable bow and teeth. |
| ancientScroll | Partly unrolled parchment. Retained rolls, writing, and seal. |
| chainmailChestplate | Laid-out mail shirt. Tapered the torso, angled the sleeves, staggered the rings, and added a collar opening detail. |
| ironLeggings | Laid-out leg armor. Tapered the legs and rounded the knee plates. |
| leatherBoots | Pair of leather boots. Rounded the toes and tapered the shafts. |
| woodenShield | Round wooden shield with metal rim and boss. Retained the readable construction. |
| unknownItem | Purple diagnostic placeholder. Retained intentionally. |

## Animation and effect coverage

- Human `idle`, `bowIdle`, `staffIdle`: retain restrained breathing; staff idle now includes the head/free-arm motion.
- Human `run`, `bowRun`, `staffRun`: solve each leg against a foot trajectory, lift the swinging foot, and keep the soles level. Staff grip compensates for body height; planted/swing transitions ease smoothly.
- Human `attack`: overhead wind-up, forward slash, wrist follow-through, and recovery to the starting pose.
- Human `pickup`: deeper crouch and reach, with ankle compensation.
- Human `cast`, `shoot`: retain timing and equipped poses. Bow string draw follows the same replicated phase as `shoot`, with release at two thirds of the clip.
- Human `chop`: retain the contact timing and axe trajectory, with restrained torso/head/free-arm counter-motion.
- Human `fishWait`, `fishAction`: rod angles toward the water; action starts and ends at the waiting pose.
- Rat `idle`, `run`, `attack`: inspected sniffing, diagonal gait, lunge, and tail motion. Softened bobbing; retained the lunge timing.
- Tree `hit`: begins at rest and damps back to rest. Damage/depletion remains driven by server state.
- Door `open`, chest `open`: inspected start, intermediate, and final hinge poses.
- Fishing spot `idle`, bow `draw`: new presentation clips; neither changes authoritative gameplay state.
- Arrow flight: follows the arc tangent, including the impact endpoint. Magic bolt spin derives from progress rather than frame count.
- Wood/stone walls: added plank seams/mortar courses and corrected linear-to-sRGB conversion in generated textures. Chat and combat text remain existing UI overlays.

## Repeatable validation

Run from `client/`:

```sh
pnpm run build
pnpm run model:check
pnpm run model:screenshot --all
pnpm run model:screenshot --animations
pnpm run model:screenshot --model human --equipment ironSword --animation attack --phase 0.58
```

`model:check` checks finite geometry and transforms across the registry, joint names, loop endpoints, seek/playback agreement, run foot clearance, action recovery, fishing pose joins, pickup reach, equipment sockets, chest state transitions, frame-independent projectiles, and disposal of owned geometry/materials. `--animations` captures every registered clip at five phases in addition to the gallery. Generated previews stay under ignored `.model-previews/`.

Visual inspection covers the complete before/after galleries, sampled animation sequences, equipped poses, hairstyles, tree damage/stump variants, open doors/chests, and rectangular buildings. A local Go-server/browser smoke check verifies registration, world/chunk/component messages, and rendering without page errors. No server, schema, authored-content, or inventory-icon formats change in this pass; armor items that previously lacked equipped body meshes retain that behavior.
