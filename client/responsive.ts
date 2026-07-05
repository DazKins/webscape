export type ViewportSize = {
  width: number;
  height: number;
};

export type DeviceProfile = ViewportSize & {
  isCoarsePointer: boolean;
  canHover: boolean;
  isPortrait: boolean;
  isMobileLayout: boolean;
};

const MOBILE_LAYOUT_WIDTH = 820;
const MOBILE_LAYOUT_HEIGHT = 560;

export function getElementSize(element: HTMLElement): ViewportSize {
  const rect = element.getBoundingClientRect();
  return {
    width: Math.max(1, Math.round(rect.width || window.innerWidth)),
    height: Math.max(1, Math.round(rect.height || window.innerHeight)),
  };
}

export function getDeviceProfile(size: ViewportSize): DeviceProfile {
  const isCoarsePointer = window.matchMedia("(pointer: coarse)").matches;
  const canHover = window.matchMedia("(hover: hover)").matches;
  const isPortrait = size.height >= size.width;
  const isMobileLayout =
    isCoarsePointer || size.width <= MOBILE_LAYOUT_WIDTH || size.height <= MOBILE_LAYOUT_HEIGHT;

  return {
    ...size,
    isCoarsePointer,
    canHover,
    isPortrait,
    isMobileLayout,
  };
}

export function subscribeToDeviceProfile(
  element: HTMLElement,
  callback: (profile: DeviceProfile) => void
) {
  const publish = () => callback(getDeviceProfile(getElementSize(element)));
  const resizeObserver = new ResizeObserver(publish);
  resizeObserver.observe(element);

  const pointerQuery = window.matchMedia("(pointer: coarse)");
  const hoverQuery = window.matchMedia("(hover: hover)");
  pointerQuery.addEventListener("change", publish);
  hoverQuery.addEventListener("change", publish);
  window.addEventListener("orientationchange", publish);
  publish();

  return () => {
    resizeObserver.disconnect();
    pointerQuery.removeEventListener("change", publish);
    hoverQuery.removeEventListener("change", publish);
    window.removeEventListener("orientationchange", publish);
  };
}
