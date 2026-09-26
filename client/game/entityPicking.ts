import * as THREE from "three";

// CSS pixels keep the extra tolerance consistent across zoom and device pixel ratios.
const PICK_PADDING_PX = 12;

export function pickEntityAtScreenPoint(
  clientX: number,
  clientY: number,
  camera: THREE.Camera,
  viewport: { width: number; height: number },
  objects: THREE.Object3D[],
): string | null {
  const raycaster = new THREE.Raycaster();
  const pick = (x: number, y: number) => {
    raycaster.setFromCamera(new THREE.Vector2(
      (x / viewport.width) * 2 - 1,
      -(y / viewport.height) * 2 + 1,
    ), camera);
    const hit = raycaster.intersectObjects(objects, true)[0];
    let object: THREE.Object3D | null = hit?.object ?? null;
    while (object && !object.userData.entityId) object = object.parent;
    return object ? { id: object.userData.entityId as string, distance: hit.distance } : null;
  };

  // Exact hits always win. Only sample nearby geometry when the pointer misses.
  const direct = pick(clientX, clientY);
  if (direct) return direct.id;
  for (const radius of [PICK_PADDING_PX / 2, PICK_PADDING_PX]) {
    let closest: { id: string; distance: number } | null = null;
    for (let step = 0; step < 16; step++) {
      const angle = step * Math.PI * 2 / 16;
      const hit = pick(clientX + Math.cos(angle) * radius, clientY + Math.sin(angle) * radius);
      if (hit && (!closest || hit.distance < closest.distance)) closest = hit;
    }
    if (closest) return closest.id;
  }
  return null;
}
