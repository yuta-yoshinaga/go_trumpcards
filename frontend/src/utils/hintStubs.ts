/**
 * Detection of hint factories that cannot return a hint.
 *
 * `hintFactories` counts a game as covered as soon as a key exists, so 22 games
 * registered as `() => null` or delegating to a `return null;` stub were being
 * reported as hinted (#4602). Both the build guard
 * (`scripts/check-hint-coverage.mjs`) and the doc-count test need the same
 * judgement, and the reviewer on that PR pointed out that two hand-synced
 * copies of the heuristic is the same drift-through-indirection the fix was
 * about, one layer down. So it lives here once.
 */

/** Matches an exported factory and captures its name and body. */
const FACTORY = /^export function (\w+)\([^)]*\)\s*:\s*HintResult \| null \{\n([\s\S]*?)\n\}/gm;

/** Matches a registration written inline as `() => null`. */
const INLINE_NULL = /^\(_?\w*\)\s*=>\s*null$/;

/** Captures the factory a registration delegates to, e.g. `getFooHint(s as FooResponse)`. */
const DELEGATE = /(\w+)\(\w+ as/;

/**
 * Names of factories whose whole body is `return null;`.
 *
 * The signature is not always `get<Game>Hint` — `chinesepokerHint` has no
 * prefix, and assuming one is how the first count of this missed it.
 *
 * @param sources - Factory file contents, keyed by anything (path or name).
 */
export function stubbedFactoryNames(sources: Record<string, string>): Set<string> {
  const stubs = new Set<string>();
  for (const [key, src] of Object.entries(sources)) {
    if (key.endsWith('.test.ts')) continue;
    for (const m of src.matchAll(FACTORY)) {
      if (m[2].trim() === 'return null;') stubs.add(m[1]);
    }
  }
  return stubs;
}

/**
 * Splits the `hintFactories` body into games that can produce a hint and games
 * whose registration always yields null.
 *
 * @param body - Source text of the `hintFactories` object literal.
 * @param stubs - Result of {@link stubbedFactoryNames}.
 */
export function splitRegistrations(body: string, stubs: Set<string>): { hinted: Set<string>; stubbed: Set<string> } {
  const hinted = new Set<string>();
  const stubbed = new Set<string>();
  for (const m of body.matchAll(/^ {2}([a-z0-9]+): (.+),$/gm)) {
    const [, game, expr] = m;
    const delegate = expr.match(DELEGATE)?.[1];
    if (INLINE_NULL.test(expr) || (delegate && stubs.has(delegate))) {
      stubbed.add(game);
    } else {
      hinted.add(game);
    }
  }
  return { hinted, stubbed };
}
