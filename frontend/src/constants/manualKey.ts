/**
 * Maps a game route path to the basename its manual Markdown file uses.
 *
 * BlackJack is served from `/` rather than `/blackjack`, so it is the one
 * route whose filename is not just the path minus its leading slash. Verified
 * against the previous hand-written 219-entry maps: every other route matched
 * this rule exactly, in both the web and CUI corpora.
 */
export function manualKeyFromPath(gamePath: string): string {
  return gamePath === '/' ? 'blackjack' : gamePath.slice(1);
}
