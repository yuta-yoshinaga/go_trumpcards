/**
 * Lazy loaders for the game manual Markdown files (Japanese, web version).
 *
 * These used to be 219 static `?raw` imports collected into one object
 * literal. Every game page reaches this module, so that put all 219 manuals —
 * plus the 219 CUI variants from the sibling module — into a single chunk
 * shared by 229 others: opening one game downloaded every game's manual
 * (~570 kB gzipped) whether or not the reader ever opened the manual.
 *
 * A non-eager `import.meta.glob` gives each manual its own chunk instead, so
 * exactly one is fetched, and only when the modal opens.
 */
import { manualKeyFromPath } from './manualKey';

/** One chunk per manual; the loader runs on first call, then Vite caches it. */
const manualModules = import.meta.glob<string>('../../../docs/manual/web/*.md', {
  query: '?raw',
  import: 'default',
});

/**
 * Fetches the web manual for a game route, or `''` when the route has none.
 *
 * The empty string preserves the caller's previous `?? ''` behaviour: an
 * unknown route renders an empty manual rather than throwing inside a modal.
 */
export async function loadManualText(gamePath: string): Promise<string> {
  const loader = manualModules[`../../../docs/manual/web/${manualKeyFromPath(gamePath)}.md`];
  return loader ? await loader() : '';
}
