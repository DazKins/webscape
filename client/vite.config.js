import { execFileSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

function getBuildRevision() {
  const override = process.env.WEBSCAPE_BUILD_REVISION?.trim();
  if (override) {
    return { revision: override, dirty: false };
  }

  const cwd = fileURLToPath(new URL(".", import.meta.url));
  const git = (...args) => execFileSync("git", args, {
    cwd,
    encoding: "utf8",
    stdio: ["ignore", "pipe", "ignore"],
  }).trim();

  try {
    return {
      revision: git("rev-parse", "HEAD"),
      dirty: git("status", "--porcelain", "--untracked-files=normal") !== "",
    };
  } catch {
    return { revision: "unknown", dirty: false };
  }
}

const build = getBuildRevision();

export default defineConfig({
  plugins: [react()],
  define: {
    __BUILD_REVISION__: JSON.stringify(build.revision),
    __BUILD_DIRTY__: JSON.stringify(build.dirty),
  },
});
