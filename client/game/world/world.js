class World {
  constructor(scene, sizeX, sizeY, input) {
    console.log("Creating world", sizeX, sizeY);
    
    this.sizeX = sizeX;
    this.sizeY = sizeY;
    this.input = input;
    
    // Create the ground plane
    this.mesh = new THREE.Mesh(
      new THREE.PlaneGeometry(sizeX, sizeY),
      new THREE.MeshBasicMaterial({ 
        color: 0x4CAF50,
        side: THREE.DoubleSide
      })
    );
    this.mesh.rotation.x = -Math.PI / 2;
    scene.add(this.mesh);

    const gridSize = sizeX;
    const divisions = sizeX;
    const gridHelper = new THREE.GridHelper(gridSize, divisions, 0x000000, 0x000000);
    
    gridHelper.material.linewidth = 1;
    gridHelper.material.depthWrite = false;
    gridHelper.position.y = 0.01;
    
    this.grid = gridHelper;
    scene.add(this.grid);

    this.highlightMesh = new THREE.Mesh(
      new THREE.PlaneGeometry(1, 1),
      new THREE.MeshBasicMaterial({
        color: 0x388E3C,
        transparent: true,
        opacity: 0.5,
        side: THREE.DoubleSide
      })
    );
    this.highlightMesh.rotation.x = -Math.PI / 2;
    this.highlightMesh.position.y = 0.02;
    scene.add(this.highlightMesh);
  }

  getHoveredTile(camera) {
    const mouse = this.input.getMousePosition();
    const mouseX = (mouse.x / window.innerWidth) * 2 - 1;
    const mouseY = -(mouse.y / window.innerHeight) * 2 + 1;

    const raycaster = new THREE.Raycaster();
    raycaster.setFromCamera(new THREE.Vector2(mouseX, mouseY), camera);

    const intersects = raycaster.intersectObject(this.mesh);

    if (intersects.length > 0) {
      const point = intersects[0].point;

      const gridX = Math.floor(point.x);
      const gridY = Math.floor(point.z);

      return { x: gridX, y: gridY };
    }
  }

  update(camera) {
    const hoveredTile = this.getHoveredTile(camera);
    if (hoveredTile) {
      this.highlightMesh.position.x = hoveredTile.x + 0.5;
      this.highlightMesh.position.z = hoveredTile.y + 0.5;
    }
  }
}

export default World;
