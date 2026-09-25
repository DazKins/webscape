import { construct, type ConstructionRequest } from "./construction";
import { geometryTransfers } from "./geometryData";

self.onmessage = (event: MessageEvent<{ id: number; request: ConstructionRequest }>) => {
  const { id, request } = event.data;
  try {
    const result = construct(request);
    const geometries = result.kind === "warmModels" ? result.geometries.map(([, data]) => data)
      : [result.surfaces.details, result.surfaces.terrain, result.surfaces.water, ...result.surfaces.wallGeometries.map(([, data]) => data)];
    self.postMessage({ id, result }, { transfer: geometries.flatMap(geometryTransfers) });
  } catch (error) {
    self.postMessage({ id, error: String(error) });
  }
};
