const BADGE_TOKEN = /\bbg-ds-(warning|info|success|error)\/\d+\b/g;

/** Opening JSX tag starting at `start`, including its closing `>`. */
function openingTag(text, start) {
  let braceDepth = 0;
  let quote = null;
  for (let i = start; i < text.length; i += 1) {
    const ch = text[i];
    if (quote) {
      if (ch === quote && text[i - 1] !== '\\') quote = null;
    } else if (ch === '"' || ch === "'") {
      quote = ch;
    } else if (ch === '{') {
      braceDepth += 1;
    } else if (ch === '}') {
      braceDepth -= 1;
    } else if (ch === '>' && braceDepth === 0) {
      return text.slice(start, i + 1);
    }
  }
  return text.slice(start);
}

/** Return the next JSX tag, respecting `>` inside quoted attributes and braces. */
function nextTag(text, start) {
  const open = text.indexOf('<', start);
  if (open === -1) return null;
  const tag = openingTag(text, open);
  if (!tag.endsWith('>')) return null;
  return { start: open, end: open + tag.length, text: tag };
}

/** Find the end of the JSX element whose opening tag starts at `start`. */
function elementEnd(text, start) {
  const first = nextTag(text, start);
  if (!first) return text.length;
  const name = /^<([A-Za-z][\w.-]*)\b/.exec(first.text)?.[1];
  if (!name || /\/\s*>$/.test(first.text)) return first.end;

  const stack = [name];
  let cursor = first.end;
  while (stack.length > 0) {
    const tag = nextTag(text, cursor);
    if (!tag) return text.length;
    cursor = tag.end;
    const closing = /^<\/([A-Za-z][\w.-]*)\s*>$/.exec(tag.text);
    if (closing && stack.at(-1) === closing[1]) {
      stack.pop();
      continue;
    }
    const nested = /^<([A-Za-z][\w.-]*)\b/.exec(tag.text)?.[1];
    if (nested && !/\/\s*>$/.test(tag.text)) stack.push(nested);
  }
  return cursor;
}

/** Return all className value ranges in JSX source. */
function classNameValues(text) {
  const out = [];
  for (const match of text.matchAll(/className=/g)) {
    const start = match.index + match[0].length;
    if (text[start] === '"' || text[start] === "'") {
      const end = text.indexOf(text[start], start + 1);
      if (end !== -1) out.push({ start, end: end + 1 });
      continue;
    }
    if (text[start] !== '{') continue;
    let depth = 0;
    for (let i = start; i < text.length; i += 1) {
      if (text[i] === '{') depth += 1;
      else if (text[i] === '}') {
        depth -= 1;
        if (depth === 0) {
          out.push({ start, end: i + 1 });
          break;
        }
      }
    }
  }
  return out;
}

/** Find semantic foregrounds that sit on matching translucent semantic backgrounds. */
export function findBadgeContrastViolations(text) {
  const violations = [];
  for (const { start, end } of classNameValues(text)) {
    const value = text.slice(start, end);
    const tagStart = text.lastIndexOf('<', start);
    const tagEnd = tagStart === -1 ? -1 : openingTag(text, tagStart).length + tagStart;
    const elementEndIndex = tagStart === -1 ? end : elementEnd(text, tagStart);
    for (const match of value.matchAll(BADGE_TOKEN)) {
      const kind = match[1];
      const foreground = new RegExp(String.raw`\btext-ds-${kind}\b`);
      const element = text.slice(tagStart, elementEndIndex);
      if (!foreground.test(element)) continue;
      violations.push({ start: tagStart, match: match[0], kind, elementEnd: elementEndIndex, tagEnd });
    }
  }
  return violations;
}
