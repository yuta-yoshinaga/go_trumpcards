/** Return quoted class-string fragments that combine a colored action background with white text. */
export function findColoredActionTextViolations(source) {
  const violations = [];
  const seen = new Set();
  const templates = [...source.matchAll(/`((?:\\.|[^`])*)`/g)];
  const addViolation = (index, match, templateIndex = index) => {
    const background = match.match(/\bbg-ds-(?:warning|accent|info|success)(?:-[\w-]+|\/\d+)?\b/)?.[0];
    const key = `${templateIndex}:${background}:text-white`;
    if (seen.has(key)) return;
    seen.add(key);
    violations.push({ index, match });
  };
  const fragments = /(['"])((?:\\.|(?!\1)[\s\S])*?)\1/g;
  for (const match of source.matchAll(fragments)) {
    if (hasViolation(match[2])) {
      const template = templates.find((item) => item.index < match.index && item.index + item[0].length > match.index);
      addViolation(match.index, match[2], template?.index ?? match.index);
    }
  }

  // A template's static text is split at each interpolation so separate class
  // branches cannot be combined into one match.
  for (const template of templates) {
    const staticParts = template[1].split(/\$\{[^{}]*\}/);
    let offset = 0;
    for (const part of staticParts) {
      if (hasViolation(part)) addViolation(template.index + 1 + offset, part, template.index);
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
