import * as THREE from "three";
import Entity from "../entity/entity";
import PositionedEntityRenderer from "./positionedEntityRenderer";

export default class RendererRock extends PositionedEntityRenderer {
  constructor(scene: THREE.Scene, entity: Entity) {
    super(scene, entity);

    const rock = new THREE.Mesh(
      new THREE.DodecahedronGeometry(0.42, 0),
      new THREE.MeshPhongMaterial({ color: 0x8b9296 })
    );
    rock.scale.set(1.1, 0.65, 0.9);
    rock.position.set(0.5, 0.28, 0.5);
    this.mesh.add(rock);

    this.addToScene();
  }
}
