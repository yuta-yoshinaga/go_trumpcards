import { type ComponentType, lazy } from 'react';

/** Async loader for one ESM module, returned by `import.meta.glob`. */
export type ModuleLoader = () => Promise<Record<string, ComponentType<unknown>>>;

/**
 * Pulls the named export `<page>Page` out of an already-imported module
 * record. Throws if the export is missing — exposed separately so the failure
 * path is unit-testable without relying on React's lazy-suspense plumbing.
 */
export function pickPageExport(
  module: Record<string, ComponentType<unknown>>,
  importPath: string,
  page: string,
): ComponentType {
  const exportName = `${page}Page`;
  const Component = module[exportName];
  if (!Component) {
    throw new Error(`gameRoutes: ${importPath} has no export named ${exportName}`);
  }
  return Component as ComponentType;
}

/**
 * Resolves a `pages/<page>Page.tsx` module from a glob result and returns a
 * `lazy()` component bound to the named export `<page>Page`.
 *
 * Extracted from `App.tsx` so the failure paths (missing module, missing named
 * export) can be unit-tested in isolation. The missing-module check throws
 * synchronously so a typo in `gameRoutes.page` surfaces at startup rather
 * than as a silent runtime crash on first navigation. The named-export check
 * lives in {@link pickPageExport} and is re-thrown inside the lazy callback;
 * React surfaces it through the nearest error boundary on first visit.
 */
export function resolvePageComponent(
  modules: Record<string, ModuleLoader>,
  routePath: string,
  page: string,
): ComponentType {
  const importPath = `./pages/${page}Page.tsx`;
  const importer = modules[importPath];
  if (!importer) {
    throw new Error(`gameRoutes: no module at ${importPath} for path "${routePath}"`);
  }
  return lazy(async () => {
    const m = await importer();
    return { default: pickPageExport(m, importPath, page) };
  });
}
