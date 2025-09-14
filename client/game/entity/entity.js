class Entity {
  constructor(id, scene) {
    this.id = id;
    this.scene = scene;

    // Create a group to hold all parts of the entity
    this.mesh = new THREE.Group();

    // Create body (cylinder)
    const bodyGeometry = new THREE.CylinderGeometry(0.3, 0.3, 0.8, 16);
    const bodyMaterial = new THREE.MeshPhongMaterial({
      color: 0x00ff00,
      shininess: 30,
    });
    const body = new THREE.Mesh(bodyGeometry, bodyMaterial);
    body.position.y = 0.4; // Position the body slightly above ground
    this.mesh.add(body);

    // Create head (sphere)
    const headGeometry = new THREE.SphereGeometry(0.25, 16, 16);
    const headMaterial = new THREE.MeshPhongMaterial({
      color: 0xffe0bd,
      shininess: 30,
    });
    const head = new THREE.Mesh(headGeometry, headMaterial);
    head.position.y = 0.9; // Position the head on top of the body
    this.mesh.add(head);

    this.scene.add(this.mesh);

    this.ticksSinceUpdate = 0;
  }

  update() {
    this.mesh.position.x =
      this.positionX + 0.5 + 0.02 * this.velocityX * this.ticksSinceUpdate;
    this.mesh.position.z =
      this.positionY + 0.5 + 0.02 * this.velocityY * this.ticksSinceUpdate;
    this.ticksSinceUpdate++;
  }

  handleEntityUpdate(entityUpdate) {
    this.ticksSinceUpdate = 0;

    this.positionX = entityUpdate.positionX;
    this.positionY = entityUpdate.positionY;
    this.velocityX = entityUpdate.velocityX;
    this.velocityY = entityUpdate.velocityY;

    // Update color only for the body (first child)
    this.mesh.children[0].material.color.set(entityUpdate.color);
  }

  remove() {
    console.log("Removing entity from scene", this.id);
    this.scene.remove(this.mesh);
  }
}

export default Entity;
