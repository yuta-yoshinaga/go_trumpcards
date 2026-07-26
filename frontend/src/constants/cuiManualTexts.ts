/**
 * Lazy loaders for the game manual Markdown files (Japanese, CUI version).
 *
 * Used when the page has CLI mode enabled so the manual matches the terminal
 * UI. Loaded per game rather than all at once — see {@link loadManualText} in
 * the sibling module for why the previous 219 static imports were a problem.
 */
import { manualKeyFromPath } from './manualKey';

/** One chunk per manual; the loader runs on first call, then Vite caches it. */
const cuiManualModules = import.meta.glob<string>('../../../docs/manual/cui/*.md', {
  query: '?raw',
  import: 'default',
});

/**
 * Fetches the CUI manual for a game route, or `''` when the route has none.
 */
export async function loadCuiManualText(gamePath: string): Promise<string> {
  const loader = cuiManualModules[`../../../docs/manual/cui/${manualKeyFromPath(gamePath)}.md`];
  return loader ? await loader() : '';
}

/** Whether the player has switched this game's page into CLI mode. */
export function isCliModeEnabled(gamePath: string): boolean {
  try {
    return localStorage.getItem(`cli-mode-${manualKeyFromPath(gamePath)}`) === 'true';
  } catch {
    return false;
  }
}
