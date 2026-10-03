/** Count Tailwind black background utilities with numeric opacity suffixes. */
export function countBgBlackUtilities(source) {
  return [...source.matchAll(/\bbg-black\/\d+\b/g)].length;
}

/** Return whether the observed count stays at or below the frozen ceiling. */
export function checkBgBlackRatchet(count, ceiling) {
  return count <= ceiling;
}
