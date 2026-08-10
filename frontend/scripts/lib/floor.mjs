/**
 * Sanity floors for guard scripts.
 *
 * Every guard in this directory walks something -- a directory tree, a route table, a JSON
 * map -- and then reports on what it found. The walk is the part that breaks silently: a
 * renamed directory, a tightened regex, or a glob that no longer matches leaves the guard
 * finding nothing, finding nothing means finding no violations, and finding no violations
 * exits 0. The build goes green and the line in the log reads exactly like coverage.
 *
 * This has already happened here: `check-dependency-licenses.mjs` inspected 341 packages and
 * reported 1, and nothing about the run looked wrong. Three guards grew ad-hoc `MIN_*`
 * constants afterwards; this module is those constants made shared and mandatory.
 * `check-guard-floors.mjs` fails the build for any guard that reports a discovered count
 * without asserting a floor under it.
 *
 * A floor is not a target. Set it well below today's real count -- low enough that ordinary
 * churn never trips it, high enough that a walk finding a small fraction of the truth does.
 * Roughly two thirds of the current count is a good default.
 */

/**
 * Assert that a walk found a plausible number of things, and exit 1 if it did not.
 *
 * @param {string} label - The guard's name, used in the failure message (e.g. `'design-tokens'`).
 * @param {number} count - How many items the walk actually found.
 * @param {number} min - The floor. Must be a positive integer; a floor of 0 asserts nothing.
 * @param {string} [what] - What is being counted, for the message (e.g. `'source files'`).
 * @returns {number} `count`, so the call can wrap an expression.
 */
export function assertFloor(label, count, min, what = 'items') {
  if (!Number.isInteger(min) || min <= 0) {
    console.error(`${label}: floor must be a positive integer, got ${min}. A floor of 0 asserts nothing.`);
    process.exit(1);
  }
  if (!Number.isFinite(count) || count < min) {
    console.error(
      `\n${label}: only ${count} ${what} found (expected at least ${min}).\n` +
        'The walk itself is broken -- it is reporting no violations because it inspected\n' +
        'almost nothing, not because there are none. Fix the walk before trusting this guard.\n',
    );
    process.exit(1);
  }
  return count;
}
