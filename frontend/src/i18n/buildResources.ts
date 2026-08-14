/**
 * Namespaces that the eager glob should skip at runtime. These bundles
 * are loaded on demand by their owning page (currently only `discover`,
 * which mounts on `/discover` and is ~25-35 KB gzipped — not worth
 * carrying for users who never open the concierge).
 */
const LAZY_NAMESPACES: ReadonlySet<string> = new Set<string>(['discover']);

/** Options for the resource builder. */
export interface BuildResourcesOptions {
  /**
   * If `true` (default), namespaces listed in `LAZY_NAMESPACES` are
   * dropped from the eager-loaded resource map. Tests pass `false` so
   * every namespace is available without a dynamic import in jsdom.
   */
  readonly skipLazy?: boolean;
}

/**
 * Extract namespace name from glob path (e.g. `./locales/ja/blackjack.json` → `blackjack`)
 * and return the resources map. By default lazy namespaces are filtered
 * out so the runtime bundle stays slim; pass `{ skipLazy: false }` to
 * include them (test setup uses this).
 */
export function buildResources(
  modules: Record<string, Record<string, string>>,
  options: BuildResourcesOptions = {},
): Record<string, Record<string, string>> {
  const { skipLazy = true } = options;
  const resources: Record<string, Record<string, string>> = {};
  for (const [path, mod] of Object.entries(modules)) {
    const name = path.split('/').pop()?.replace('.json', '');
    if (!name) continue;
    if (skipLazy && LAZY_NAMESPACES.has(name)) continue;
    resources[name] = mod;
  }
  return resources;
}
