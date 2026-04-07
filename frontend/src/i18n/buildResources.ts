/** Extract namespace name from glob path (e.g. './locales/ja/blackjack.json' → 'blackjack'). */
export function buildResources(
  modules: Record<string, Record<string, string>>,
): Record<string, Record<string, string>> {
  const resources: Record<string, Record<string, string>> = {};
  for (const [path, mod] of Object.entries(modules)) {
    const name = path.split('/').pop()?.replace('.json', '');
    if (name) resources[name] = mod;
  }
  return resources;
}
