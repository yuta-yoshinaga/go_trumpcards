/** Finds inline expressions that add a plus sign to non-negative values.
 * Does not detect n < 0 ? … : '+' ordering, Math.sign, or semicolon-separated multi-line ternary chains.
 */
export function findInlineSignedDelta(text) {
  return [...text.matchAll(/(?:>\s*0|>=\s*0)\s*\?\s*(?:`\+\$\{|['"]\+['"])/g)]
    .filter((match) => {
      const conditionValue = text.slice(0, match.index).match(/([\w.]+)\s*$/)?.[1];
      const interpolatedValue = text.slice(match.index + match[0].length).match(/^([^}]+)\}/)?.[1];
      if (interpolatedValue && conditionValue !== interpolatedValue) return false;
      const semicolon = text.indexOf(';', match.index);
      const expression = text.slice(match.index, semicolon < 0 ? match.index + 300 : semicolon);
      return !/:\s*[\w.]+\s*<\s*0\s*\?/.test(expression);
    })
    .map((match) => ({ index: match.index, match: match[0] }));
}
