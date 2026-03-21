// @ts-expect-error no types for jsdom
import { JSDOM } from 'jsdom';

const dom = new JSDOM('<!DOCTYPE html><html><body></body></html>', {
  url: 'http://localhost',
  pretendToBeVisual: true,
});

const win = dom.window as unknown as Record<string, unknown>;
const g = globalThis as unknown as Record<string, unknown>;

// Keys that must NOT be overridden (Bun runtime essentials)
const skipKeys = new Set([
  'globalThis',
  'global',
  'process',
  'console',
  'Buffer',
  'Bun',
  'module',
  'exports',
  'require',
  '__dirname',
  '__filename',
  'queueMicrotask',
  'structuredClone',
  'crypto',
  'fetch',
  'Request',
  'Response',
  'Headers',
  'TextEncoder',
  'TextDecoder',
  'ReadableStream',
  'WritableStream',
  'TransformStream',
  'setTimeout',
  'clearTimeout',
  'setInterval',
  'clearInterval',
  'setImmediate',
  'clearImmediate',
  'URL',
  'URLSearchParams',
  'Blob',
  'File',
  'FormData',
  'AbortController',
  'AbortSignal',
  'atob',
  'btoa',
  'performance',
]);

// Copy all jsdom window properties to globalThis
for (const key of Object.getOwnPropertyNames(win)) {
  if (skipKeys.has(key)) continue;
  try {
    g[key] = win[key];
  } catch {
    // Some properties may be readonly
  }
}

// Override key globals that tests depend on (use defineProperty for readonly ones)
const overrides: Record<string, unknown> = {
  document: win.document,
  navigator: win.navigator,
  HTMLElement: win.HTMLElement,
  getComputedStyle: win.getComputedStyle,
  localStorage: win.localStorage,
  sessionStorage: win.sessionStorage,
  Storage: win.Storage,
  Event: win.Event,
  CustomEvent: win.CustomEvent,
  MouseEvent: win.MouseEvent,
  KeyboardEvent: win.KeyboardEvent,
  Node: win.Node,
  Element: win.Element,
  DocumentFragment: win.DocumentFragment,
  MutationObserver: win.MutationObserver,
  SVGElement: win.SVGElement,
  history: win.history,
  location: win.location,
  Image: win.Image,
  requestAnimationFrame: win.requestAnimationFrame,
  cancelAnimationFrame: win.cancelAnimationFrame,
};

for (const [key, value] of Object.entries(overrides)) {
  try {
    g[key] = value;
  } catch {
    try {
      Object.defineProperty(g, key, { value, writable: true, configurable: true });
    } catch {
      // Skip if truly immutable
    }
  }
}
