import { spawn } from "node:child_process";
import { mkdir } from "node:fs/promises";
import net from "node:net";
import path from "node:path";
import process from "node:process";
import { chromium } from "playwright";

const options = parseArguments(process.argv.slice(2));
const port = await availablePort();
const baseUrl = `http://127.0.0.1:${port}`;
const viteExecutable = path.resolve("node_modules", ".bin", "vite");
const server = spawn(viteExecutable, ["--host", "127.0.0.1", "--port", String(port), "--strictPort"], {
  cwd: process.cwd(),
  stdio: ["ignore", "pipe", "pipe"],
});
let serverOutput = "";
server.stdout.on("data", (chunk) => { serverOutput += chunk.toString(); });
server.stderr.on("data", (chunk) => { serverOutput += chunk.toString(); });

let browser;
try {
  await waitForServer(`${baseUrl}/model-lab.html`);
  browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 800, height: 800 }, deviceScaleFactor: 1 });

  if (options.all) {
    await openPreview(page, baseUrl, { model: "human", view: options.view });
    const models = await page.evaluate(() => window.__MODEL_LAB_MODELS__ ?? []);
    const outputDirectory = options.output ?? path.resolve(".model-previews", "gallery");
    await mkdir(outputDirectory, { recursive: true });
    for (const model of models) {
      await openPreview(page, baseUrl, { model, view: options.view });
      const output = path.join(outputDirectory, `${safeName(model)}.png`);
      await page.screenshot({ path: output });
      process.stdout.write(`${output}\n`);
    }
  } else {
    const capture = {
      model: options.model,
      equipment: options.equipment,
      animation: options.animation,
      phase: options.phase,
      damageStage: options.damageStage,
      appearance: options.appearance,
      view: options.view,
    };
    await openPreview(page, baseUrl, capture);
    const output = resolveSingleOutput(options, capture);
    await mkdir(path.dirname(output), { recursive: true });
    await page.screenshot({ path: output });
    process.stdout.write(`${output}\n`);
  }
} catch (error) {
  if (serverOutput) {
    process.stderr.write(serverOutput);
  }
  throw error;
} finally {
  await browser?.close();
  server.kill("SIGTERM");
}

async function openPreview(page, baseUrl, capture) {
  const query = new URLSearchParams({
    model: capture.model,
    view: capture.view,
    capture: "1",
  });
  if (capture.animation) {
    query.set("animation", capture.animation);
    query.set("phase", String(capture.phase));
  }
  if (capture.equipment) {
    query.set("equipment", capture.equipment);
  }
  if (capture.damageStage > 0) {
    query.set("damageStage", String(capture.damageStage));
  }
  if (capture.appearance) {
    for (const [key, value] of Object.entries(capture.appearance)) {
      query.set(key, value);
    }
  }
  await page.goto(`${baseUrl}/model-lab.html?${query}`, { waitUntil: "networkidle" });
  await page.waitForFunction(() => window.__MODEL_LAB_READY__ === true);
}

function resolveSingleOutput(options, capture) {
  if (options.output?.toLowerCase().endsWith(".png")) {
    return path.resolve(options.output);
  }
  const directory = path.resolve(options.output ?? ".model-previews");
  const parts = [capture.model];
  if (capture.equipment) {
    parts.push("with", capture.equipment);
  }
  if (capture.animation) {
    parts.push(capture.animation, `phase-${formatPhase(capture.phase)}`);
  }
  if (capture.appearance) {
    parts.push(capture.appearance.hairStyle);
  }
  parts.push(capture.view);
  return path.join(directory, `${parts.map(safeName).join("-")}.png`);
}

function parseArguments(args) {
  const result = {
    all: false,
    model: "human",
    equipment: "",
    animation: "",
    phase: 0,
    damageStage: 0,
    appearance: undefined,
    view: "three-quarter",
    output: undefined,
  };
  for (let index = 0; index < args.length; index += 1) {
    const argument = args[index];
    switch (argument) {
      case "--all":
        result.all = true;
        break;
      case "--model":
        result.model = requiredValue(args, ++index, argument);
        break;
      case "--animation":
        result.animation = requiredValue(args, ++index, argument);
        break;
      case "--equipment":
        result.equipment = requiredValue(args, ++index, argument);
        break;
      case "--phase":
        result.phase = Number(requiredValue(args, ++index, argument));
        if (!Number.isFinite(result.phase) || result.phase < 0 || result.phase > 1) {
          throw new Error("--phase must be between 0 and 1");
        }
        break;
      case "--damage-stage":
        result.damageStage = Number(requiredValue(args, ++index, argument));
        if (!Number.isInteger(result.damageStage) || result.damageStage < 0 || result.damageStage > 4) {
          throw new Error("--damage-stage must be an integer between 0 and 4");
        }
        break;
      case "--skin-tone":
      case "--hair-style":
      case "--hair-color":
      case "--tunic-color":
      case "--trousers-color":
      case "--shoe-color": {
        const keys = {
          "--skin-tone": "skinTone",
          "--hair-style": "hairStyle",
          "--hair-color": "hairColor",
          "--tunic-color": "tunicColor",
          "--trousers-color": "trousersColor",
          "--shoe-color": "shoeColor",
        };
        result.appearance ??= {
          skinTone: "fair",
          hairStyle: "cropped",
          hairColor: "darkBrown",
          tunicColor: "slateBlue",
          trousersColor: "navy",
          shoeColor: "darkBrown",
        };
        result.appearance[keys[argument]] = requiredValue(args, ++index, argument);
        break;
      }
      case "--view":
        result.view = requiredValue(args, ++index, argument);
        if (!["front", "back", "side", "top", "three-quarter"].includes(result.view)) {
          throw new Error("--view must be front, back, side, top, or three-quarter");
        }
        break;
      case "--output":
        result.output = path.resolve(requiredValue(args, ++index, argument));
        break;
      default:
        if (!argument.startsWith("-") && result.model === "human") {
          result.model = argument;
          break;
        }
        throw new Error(`unknown argument: ${argument}`);
    }
  }
  return result;
}

function requiredValue(args, index, option) {
  const value = args[index];
  if (!value || value.startsWith("--")) {
    throw new Error(`${option} requires a value`);
  }
  return value;
}

function formatPhase(value) {
  return Number(value).toFixed(2).replace(".", "-");
}

function safeName(value) {
  return String(value).toLowerCase().replace(/[^a-z0-9-]+/g, "-").replace(/^-|-$/g, "");
}

async function availablePort() {
  return await new Promise((resolve, reject) => {
    const server = net.createServer();
    server.unref();
    server.on("error", reject);
    server.listen(0, "127.0.0.1", () => {
      const address = server.address();
      const port = typeof address === "object" && address ? address.port : 4173;
      server.close(() => resolve(port));
    });
  });
}

async function waitForServer(url) {
  const deadline = Date.now() + 20_000;
  while (Date.now() < deadline) {
    if (server.exitCode !== null) {
      throw new Error(`Vite exited before it became ready (${server.exitCode})`);
    }
    try {
      const response = await fetch(url);
      if (response.ok) {
        return;
      }
    } catch {
      // Vite is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 100));
  }
  throw new Error("timed out waiting for the model lab Vite server");
}
