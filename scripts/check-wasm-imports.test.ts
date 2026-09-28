import { describe, expect, test } from "bun:test";
import { findMissingImports } from "./check-wasm-imports";

describe("findMissingImports", () => {
  test("accepts imports supplied by wasm_exec glue", () => {
    expect(findMissingImports([{ module: "gojs", name: "runtime.getRandomData", kind: "function" }],
      'gojs: { "runtime.getRandomData": () => {} }')).toEqual([]);
  });

  test("reports imports absent from wasm_exec glue", () => {
    expect(findMissingImports([{ module: "gojs", name: "runtime.getRandomData", kind: "function" }],
      'gojs: { "runtime.ticks": () => {} }')).toEqual(["gojs.runtime.getRandomData"]);
  });

  test("accepts imports supplied by a separate worker module", () => {
    const wasmExec = 'gojs: { "runtime.ticks": () => {} }';
    const worker = 'workers: { ready: () => {} }';
    expect(findMissingImports([{ module: "workers", name: "ready", kind: "function" }], `${wasmExec}\n${worker}`)).toEqual([]);
  });

  test("escapes dots in import names", () => {
    expect(findMissingImports([{ module: "runtime", name: "getRandomData", kind: "function" }],
      'runtime: { "getRandomData": () => {} } runtimeXgetRandomData: () => {}')).toEqual([]);
    expect(findMissingImports([{ module: "gojs", name: "runtime.getRandomData", kind: "function" }],
      'gojs: { "runtimeXgetRandomData": () => {} }')).toEqual(["gojs.runtime.getRandomData"]);
  });
});
