class Entity {
  constructor(id, scene) {
    this.id = id;
    this.scene = scene;

    this.mesh = new THREE.Group();
    this.mesh.userData.entityId = id;

    const bodyGeometry = new THREE.CylinderGeometry(0.3, 0.3, 0.8, 16);
    const bodyMaterial = new THREE.MeshPhongMaterial({
      color: 0x00ff00,
      shininess: 30,
    });
    const body = new THREE.Mesh(bodyGeometry, bodyMaterial);
    body.position.y = 0.4;
    this.mesh.add(body);

    const headGeometry = new THREE.SphereGeometry(0.25, 16, 16);
    const headMaterial = new THREE.MeshPhongMaterial({
      color: 0xffe0bd,
      shininess: 30,
    });
    const head = new THREE.Mesh(headGeometry, headMaterial);
    head.position.y = 0.9;
    this.mesh.add(head);

    this.scene.add(this.mesh);
  }

  update() {
    this.mesh.position.x +=
      (this.positionX + 0.5 - this.mesh.position.x) * 0.05;
    this.mesh.position.z +=
      (this.positionY + 0.5 - this.mesh.position.z) * 0.05;
  }

  handleEntityUpdate(entityUpdate) {
    const positionComponent = entityUpdate.components.position;
    const metadataComponent = entityUpdate.components.metadata;

    this.positionX = positionComponent.x;
    this.positionY = positionComponent.y;
    this.name = metadataComponent.name;

    // Update color only for the body (first child)
    this.mesh.children[0].material.color.set(metadataComponent.color);
  }

  remove() {
    console.log("Removing entity from scene", this.id);
    this.scene.remove(this.mesh);
  }
}

export default Entity;
