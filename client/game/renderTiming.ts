type TimerExtension = {
  TIME_ELAPSED_EXT: number;
  GPU_DISJOINT_EXT: number;
};

export type RenderTimingSample = {
  cpuMs: number | null;
  gpuMs: number | null;
  gpuSupported: boolean;
};

// GPU queries are resolved on later frames, never by waiting for the GPU.
export default class RenderTiming {
  private gl: WebGL2RenderingContext;
  private extension: TimerExtension | null;
  private query: WebGLQuery | null = null;
  private queryActive = false;
  private contextLost = false;
  private startedAt: number | null = null;
  private cpuTotal = 0;
  private cpuCount = 0;
  private gpuTotal = 0;
  private gpuCount = 0;

  constructor(gl: WebGL2RenderingContext) {
    this.gl = gl;
    this.extension = gl.getExtension("EXT_disjoint_timer_query_webgl2");
  }

  begin() {
    const gl = this.gl;
    if (gl.isContextLost()) {
      this.reset();
      this.contextLost = true;
      return;
    }
    if (this.contextLost) {
      this.extension = gl.getExtension("EXT_disjoint_timer_query_webgl2");
      this.contextLost = false;
    }
    const ext = this.extension;
    if (ext) {
      if (gl.getParameter(ext.GPU_DISJOINT_EXT)) {
        // A GPU clock interruption invalidates outstanding measurements.
        this.reset();
      } else {
        if (this.query && gl.getQueryParameter(this.query, gl.QUERY_RESULT_AVAILABLE)) {
          const nanoseconds = gl.getQueryParameter(this.query, gl.QUERY_RESULT) as number;
          if (Number.isFinite(nanoseconds) && nanoseconds >= 0) {
            this.gpuTotal += nanoseconds / 1e6;
            this.gpuCount++;
          }
          gl.deleteQuery(this.query);
          this.query = null;
        }
        // Keep at most one outstanding query, even if results are delayed.
        if (!this.query) {
          this.query = gl.createQuery();
          if (this.query) {
            gl.beginQuery(ext.TIME_ELAPSED_EXT, this.query);
            this.queryActive = true;
          }
        }
      }
    }
    this.startedAt = performance.now();
  }

  end() {
    if (this.startedAt !== null) {
      this.cpuTotal += performance.now() - this.startedAt;
      this.cpuCount++;
      this.startedAt = null;
    }
    if (this.queryActive && this.extension) {
      this.gl.endQuery(this.extension.TIME_ELAPSED_EXT);
      this.queryActive = false;
    }
  }

  takeSample(): RenderTimingSample {
    const sample = {
      cpuMs: this.cpuCount ? this.cpuTotal / this.cpuCount : null,
      gpuMs: this.gpuCount ? this.gpuTotal / this.gpuCount : null,
      gpuSupported: this.extension !== null,
    };
    this.cpuTotal = this.cpuCount = this.gpuTotal = this.gpuCount = 0;
    return sample;
  }

  reset() {
    if (this.queryActive && this.extension) this.gl.endQuery(this.extension.TIME_ELAPSED_EXT);
    if (this.query) this.gl.deleteQuery(this.query);
    this.query = null;
    this.queryActive = false;
    this.startedAt = null;
    this.cpuTotal = this.cpuCount = this.gpuTotal = this.gpuCount = 0;
  }
}
