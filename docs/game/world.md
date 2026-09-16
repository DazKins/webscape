# Willowbrook starting area

A village square, a western forest, and eastern training grounds, using 32 × 32 chunks. The default arrival is (15, 18), beside the village guide and DazKins. North is decreasing world Y. The road at Y 14–16 joins all three districts; the older streaming checkpoint and far plaza remain available farther east.

| District | Places and people |
| --- | --- |
| Willowbrook (0, 0) | Founders fountain and lantern-lined square; Mara’s Trail Supplies northwest; Ember & Iron northeast; Copper Kettle southwest; Lantern Archive southeast. Nessa teaches fishing beside the northern pond; Sage Lumen teaches magic outside the archive. |
| Willow Wood (-1, 0) | Rowan teaches woodcutting beside the road. Old Fen lives in the lodge to the north. A winding trail leads to a fishing pool and ruined shrine, with a wood rat near the road. |
| Eastwatch (1, 0) | Tamsin’s Bent Bow and the training yard north of the road, with Iris (ranged) and Captain Ash (melee). Field rats are the easiest opponents; quarry rats south of the road are tougher. Armed outlaws occupy the southeastern watchhouse. |

Terrain ranges from height 0 in the quarry to 10 on the wooded ridges. The village square sits at height 4, with a lower pond and inn, a raised forge and archive, and level building floors. The training yard overlooks the quarry; the abandoned watchhouse occupies a higher ridge. Both fishing pools and their immediate banks are level. Hills taper through the western edge of the streaming checkpoint so chunk transitions remain continuous.

The fourteen village NPCs wander within small areas around their authored positions. Shopkeepers, Fen, and Nessa stay within one tile; most tutors and guides roam within two, while DazKins can patrol three tiles. Initial walk timers are staggered, and the existing movement system pauses NPCs during conversations and trading.

Each tutor visibly carries the appropriate tool or weapon. Tutor conversations explain existing mechanics, supplies, and nearby practice locations; they do not award skill levels or claim an unimplemented XP system. Shops have distinct inventories and buyback prices. Decorative archery targets are scenery; field rats are the live practice targets.

Buildings use open-top wall outlines so interiors stay visible, with working doors, furniture, and loot. Most supply chests deplete once per world lifetime. The wayfarer key chest stays repeatable so subsequent players can complete DazKins’ errand. The existing quest and conversation event ids are preserved: accept the errand in the square, collect the key inside Rowan’s Lodge, then defeat a rat and return to DazKins for the Ancient Scroll. Killing the rat advances to the return objective; speaking to DazKins afterward completes the quest.

## Related documentation

- [Character bible](characters/index.md)
- [World content implementation](../technical/world-content.md): authored equipment and validation.
- [Game project manifest](../../game-project/game.json): the content files loaded by the game.
