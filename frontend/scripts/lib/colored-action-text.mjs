/** Return quoted class-string fragments that combine a colored action background with white text. */
export function findColoredActionTextViolations(source) {
  const violations = [];
  const fragments = /(['"])((?:\\.|(?!\1)[\s\S])*?)\1/g;
  for (const match of source.matchAll(fragments)) {
    if (hasViolation(match[2])) violations.push({ index: match.index, match: match[2] });
  }

  // A template's static text is split at each interpolation so separate class
  // branches cannot be combined into one match.
  const templates = /`((?:\\.|[^`])*)`/g;
  for (const template of source.matchAll(templates)) {
    const staticParts = template[1].split(/\$\{[^{}]*\}/);
    let offset = 0;
    for (const part of staticParts) {
      if (hasViolation(part)) violations.push({ index: template.index + 1 + offset, match: part });
      offset += part.length + 3;
    }
  }
  return violations.sort((a, b) => a.index - b.index);
}

function hasViolation(fragment) {
  return (
    /\bbg-ds-(?:warning|accent|info|success)(?:-[\w-]+|\/\d+)?\b/.test(fragment) && /\btext-white\b/.test(fragment)
  );
}
