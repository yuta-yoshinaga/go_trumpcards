const UTILITY_PATTERN =
  /(?:^|[^a-z0-9-])((?:(?:[a-z0-9-]+):)*(?:[a-z]+)-(ds-[a-z0-9-]+)(?:\/\d+(?:\.\d+)?)?)(?![a-z0-9-])/g;

/** Return the design-system token names declared in an @theme block. */
export function collectDesignTokens(indexCss) {
  const theme = indexCss.match(/@theme\s*\{([\s\S]*?)\}/)?.[1] ?? '';
  const tokens = { color: new Set(), shadow: new Set() };
  for (const match of theme.matchAll(/--(color|shadow)-(ds-[a-z0-9-]+)\s*:/g)) {
    tokens[match[1]].add(match[2]);
  }
  return tokens;
}

/** Find design-system utility references whose token is not declared in @theme. */
export function findUndefinedDesignUtilities(source, tokens) {
  const violations = [];
  for (const match of source.matchAll(UTILITY_PATTERN)) {
    const utility = match[1];
    const base = utility.slice(utility.lastIndexOf(':') + 1);
    const token = match[2];
    const isShadow = base.startsWith('shadow-');
    const defined = tokens.color.has(token) || (isShadow && tokens.shadow.has(token));
    if (!defined) violations.push({ utility: base.replace(/\/\d+(?:\.\d+)?$/, ''), index: match.index });
  }
  return violations;
}
