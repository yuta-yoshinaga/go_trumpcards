/** Find legacy error text utilities outside class strings with a white background. */
export function findErrorTextTokenViolations(source) {
  const violations = [];
  const classStrings = /(?:"([^"]*)"|'([^']*)'|`([^`]*)`)/g;
  for (const match of source.matchAll(classStrings)) {
    const classes = match[1] ?? match[2] ?? match[3] ?? '';
    if (!/(?:^|\s|:)text-ds-error(?:\s|$)/.test(classes) || /(?:^|\s)bg-white(?:\s|$)/.test(classes)) continue;
    violations.push({ index: match.index, classes });
  }
  return violations;
}
