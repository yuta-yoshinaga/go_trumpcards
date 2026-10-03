/** Count Tailwind black background utilities with numeric opacity suffixes. */
export function countBgBlackUtilities(source) {
  return [...source.matchAll(/\bbg-black\/(\d+|\[[^\]]+\])/g)].length;
}

/** Return whether the observed count stays at or below the frozen ceiling. */
export function checkBgBlackRatchet(count, ceiling) {
  return count <= ceiling;
}

/** Add an over-ceiling bg-black count to the shared design-token violations. */
export function addBgBlackRatchetViolation(violations, count, ceiling) {
  if (checkBgBlackRatchet(count, ceiling)) return;

  violations.push({
    file: 'src/components and src/pages',
    match: `${count} bg-black utilities`,
    message: 'Do not add new uses; lower the ceiling when existing uses are removed.',
  });
}
