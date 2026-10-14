const OPACITY_SUFFIXED_TEXT_TOKEN = /\btext-ds-[a-z]+\/\d+\b/g;

/** Finds opacity-suffixed design-system text utilities in source text. */
export function findOpacitySuffixedTextTokens(source) {
  return [...source.matchAll(OPACITY_SUFFIXED_TEXT_TOKEN)].map((match) => ({
    token: match[0],
    index: match.index,
  }));
}
