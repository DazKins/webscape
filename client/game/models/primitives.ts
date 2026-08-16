import * as THREE from "three";

type MeshOptions = {
  roughness?: number;
  metalness?: number;
  emissive?: THREE.ColorRepresentation;
};

export function mesh(
  geometry: THREE.BufferGeometry,
  color: THREE.ColorRepresentation,
  options: MeshOptions = {},
): THREE.Mesh {
  const material = new THREE.MeshStandardMaterial({
    color,
    emissive: options.emissive,
    flatShading: true,
    roughness: options.roughness ?? 0.82,
    metalness: options.metalness ?? 0,
  });
  const result = new THREE.Mesh(geometry, material);
  result.castShadow = true;
  result.receiveShadow = true;
  return result;
}

export function box(
  width: number,
  height: number,
  depth: number,
  color: THREE.ColorRepresentation,
  options?: MeshOptions,
) {
  return mesh(new THREE.BoxGeometry(width, height, depth), color, options);
}

export function taperedBox(
  topWidth: number,
  topDepth: number,
  bottomWidth: number,
  bottomDepth: number,
  height: number,
  color: THREE.ColorRepresentation,
  options?: MeshOptions,
) {
  const topX = topWidth / 2;
  const topZ = topDepth / 2;
  const bottomX = bottomWidth / 2;
  const bottomZ = bottomDepth / 2;
  const halfHeight = height / 2;
  const geometry = new THREE.BufferGeometry();
  geometry.setAttribute(
    "position",
    new THREE.Float32BufferAttribute(
      [
        -bottomX, -halfHeight, -bottomZ,
        bottomX, -halfHeight, -bottomZ,
        bottomX, -halfHeight, bottomZ,
        -bottomX, -halfHeight, bottomZ,
        -topX, halfHeight, -topZ,
        topX, halfHeight, -topZ,
        topX, halfHeight, topZ,
        -topX, halfHeight, topZ,
      ],
      3,
    ),
  );
  geometry.setIndex([
    0, 1, 2, 0, 2, 3,
    4, 7, 6, 4, 6, 5,
    0, 3, 7, 0, 7, 4,
    1, 5, 6, 1, 6, 2,
    0, 4, 5, 0, 5, 1,
    3, 2, 6, 3, 6, 7,
  ]);
  geometry.computeVertexNormals();
  return mesh(geometry, color, options);
}

export function cylinder(
  radiusTop: number,
  radiusBottom: number,
  height: number,
  segments: number,
  color: THREE.ColorRepresentation,
  options?: MeshOptions,
) {
  return mesh(
    new THREE.CylinderGeometry(radiusTop, radiusBottom, height, segments),
    color,
    options,
  );
}

export function cone(
  radius: number,
  height: number,
  segments: number,
  color: THREE.ColorRepresentation,
  options?: MeshOptions,
) {
  return mesh(new THREE.ConeGeometry(radius, height, segments), color, options);
}

export function sphere(
  radius: number,
  widthSegments: number,
  heightSegments: number,
  color: THREE.ColorRepresentation,
  options?: MeshOptions,
) {
  return mesh(
    new THREE.SphereGeometry(radius, widthSegments, heightSegments),
    color,
    options,
  );
}

export function sphereCap(
  radius: number,
  widthSegments: number,
  heightSegments: number,
  thetaLength: number,
  color: THREE.ColorRepresentation,
  options?: MeshOptions,
) {
  return mesh(
    new THREE.SphereGeometry(
      radius,
      widthSegments,
      heightSegments,
      0,
      Math.PI * 2,
      0,
      thetaLength,
    ),
    color,
    options,
  );
}

export function dodecahedron(
  radius: number,
  color: THREE.ColorRepresentation,
  options?: MeshOptions,
) {
  return mesh(new THREE.DodecahedronGeometry(radius, 0), color, options);
}

export function torus(
  radius: number,
  tube: number,
  radialSegments: number,
  tubularSegments: number,
  color: THREE.ColorRepresentation,
  options?: MeshOptions,
) {
  return mesh(
    new THREE.TorusGeometry(radius, tube, radialSegments, tubularSegments),
    color,
    options,
  );
}
