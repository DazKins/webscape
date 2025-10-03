class World {
  constructor(scene, sizeX, sizeY, walls, input) {
    console.log("Creating world", sizeX, sizeY);

    this.sizeX = sizeX;
    this.sizeY = sizeY;
    this.walls = walls;
    this.input = input;

    // Create the ground plane
    this.mesh = new THREE.Mesh(
      new THREE.PlaneGeometry(sizeX, sizeY),
      new THREE.MeshBasicMaterial({
        color: 0x4caf50,
        side: THREE.DoubleSide,
      })
    );
    this.mesh.rotation.x = -Math.PI / 2;
    scene.add(this.mesh);

    const gridSize = sizeX;
    const divisions = sizeX;
    const gridHelper = new THREE.GridHelper(
      gridSize,
      divisions,
      0x000000,
      0x000000
    );

    gridHelper.material.linewidth = 1;
    gridHelper.material.depthWrite = false;
    gridHelper.position.y = 0.01;

    this.grid = gridHelper;
    scene.add(this.grid);

    this.highlightMesh = new THREE.Mesh(
      new THREE.PlaneGeometry(1, 1),
      new THREE.MeshBasicMaterial({
        color: 0x388e3c,
        transparent: true,
        opacity: 0.5,
        side: THREE.DoubleSide,
      })
    );
    this.highlightMesh.rotation.x = -Math.PI / 2;
    this.highlightMesh.position.y = 0.02;
    scene.add(this.highlightMesh);

    this.walls.forEach((row, x) => {
      row.forEach((wall, y) => {
        if (wall === false) return;

        // Create the main wall mesh with pastel grey color
        const wallMesh = new THREE.Mesh(
          new THREE.BoxGeometry(1, 1, 1),
          new THREE.MeshBasicMaterial({ color: 0xd3d3d3 })
        );
        
        // Create wireframe for black edges
        const wireframeGeometry = new THREE.EdgesGeometry(new THREE.BoxGeometry(1, 1, 1));
        const wireframeMaterial = new THREE.LineBasicMaterial({ 
          color: 0x000000, 
          linewidth: 2,
          transparent: true,
          opacity: 1.0
        });
        const wireframe = new THREE.LineSegments(wireframeGeometry, wireframeMaterial);
        
        // Add wireframe to the wall mesh
        wallMesh.add(wireframe);
        wallMesh.position.x = x - sizeX / 2 + 0.5;
        wallMesh.position.z = y - sizeY / 2 + 0.5;
        wallMesh.position.y = 0.5;
        scene.add(wallMesh);
      });
    });
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
