import { readFile } from "node:fs/promises";
import { join } from "node:path";

export type WasmImport = { module: string; name: string; kind: string };

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

export function findMissingImports(imports: WasmImport[], glue: string): string[] {
  const missing: string[] = [];
  for (const item of imports) {
    if (item.kind !== "function") continue;
    const module = escapeRegExp(item.module);
    const name = escapeRegExp(item.name);
    const modulePattern = new RegExp(`(?:["']${module}["']|\\b${module})\\s*:\\s*\\{`);
    const moduleMatch = modulePattern.exec(glue);
    // Imports for one module can be split between worker.mjs and wasm_exec.js,
    // so intentionally search through the end of the combined glue. A later file
    // with the same key could mask a missing import, but names are namespace-qualified (e.g. runtime.getRandomData).
    const body = moduleMatch ? glue.slice(moduleMatch.index + moduleMatch[0].length) : "";
    const namePattern = new RegExp(`(?:["']${name}["']|\\b${name})\\s*:`);
    if (!namePattern.test(body)) missing.push(`${item.module}.${item.name}`);
  }
  return missing;
}

async function main(): Promise<void> {
  const buildDir = process.argv[2];
  if (!buildDir) throw new Error("usage: bun scripts/check-wasm-imports.ts workers/<worker>/build");
  const files = (await Array.fromAsync(new Bun.Glob("*").scan({ cwd: buildDir })))
    .filter((file) => file.endsWith(".js") || file.endsWith(".mjs"));
  if (!files.length) throw new Error(`no .js or .mjs files found in ${buildDir}`);
  const glue = await Promise.all(files.map((file) => readFile(join(buildDir, file), "utf8"))).then((parts) => parts.join("\n"));
  const wasm = await readFile(join(buildDir, "app.wasm"));
  const imports = WebAssembly.Module.imports(await WebAssembly.compile(wasm));
  const missing = findMissingImports(imports, glue);
  if (missing.length) {
    console.error(`Missing wasm imports in ${buildDir}:\n${missing.map((name) => `  ${name}`).join("\n")}`);
    process.exitCode = 1;
    return;
  }
  console.log(`All ${imports.length} wasm imports are provided by ${files.join(", ")}`);
}

if (import.meta.main) await main();
