import * as THREE from "three";
import { sharedGeometry, sharedMaterial } from "./assetCache";

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
  const materialParameters: THREE.MeshStandardMaterialParameters = {
    color,
    flatShading: true,
    roughness: options.roughness ?? 0.82,
    metalness: options.metalness ?? 0,
  };
  if (options.emissive !== undefined) {
    materialParameters.emissive = options.emissive;
  }
  const material = sharedMaterial(JSON.stringify({
    ...materialParameters,
    color: new THREE.Color(color).toArray(),
    emissive: options.emissive === undefined ? 0 : new THREE.Color(options.emissive).toArray(),
  }), () => new THREE.MeshStandardMaterial(materialParameters));
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
  return mesh(sharedGeometry(`box:${width}:${height}:${depth}`, () => new THREE.BoxGeometry(width, height, depth)), color, options);
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
  const geometry = sharedGeometry(`taperedBox:${topWidth}:${topDepth}:${bottomWidth}:${bottomDepth}:${height}`, () => {
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
    return geometry;
  });
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
    sharedGeometry(`cylinder:${radiusTop}:${radiusBottom}:${height}:${segments}`, () => new THREE.CylinderGeometry(radiusTop, radiusBottom, height, segments)),
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
  return mesh(sharedGeometry(`cone:${radius}:${height}:${segments}`, () => new THREE.ConeGeometry(radius, height, segments)), color, options);
}

export function sphere(
  radius: number,
  widthSegments: number,
  heightSegments: number,
  color: THREE.ColorRepresentation,
  options?: MeshOptions,
) {
  return mesh(
    sharedGeometry(`sphere:${radius}:${widthSegments}:${heightSegments}`, () => new THREE.SphereGeometry(radius, widthSegments, heightSegments)),
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
    sharedGeometry(`sphereCap:${radius}:${widthSegments}:${heightSegments}:${thetaLength}`, () => new THREE.SphereGeometry(
      radius,
      widthSegments,
      heightSegments,
      0,
      Math.PI * 2,
      0,
      thetaLength,
    )),
    color,
    options,
  );
}

export function dodecahedron(
  radius: number,
  color: THREE.ColorRepresentation,
  options?: MeshOptions,
) {
  return mesh(sharedGeometry(`dodecahedron:${radius}`, () => new THREE.DodecahedronGeometry(radius, 0)), color, options);
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
    sharedGeometry(`torus:${radius}:${tube}:${radialSegments}:${tubularSegments}`, () => new THREE.TorusGeometry(radius, tube, radialSegments, tubularSegments)),
    color,
    options,
  );
}
