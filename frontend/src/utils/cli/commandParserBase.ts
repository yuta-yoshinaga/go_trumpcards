/** Split user input into a lowercased command and string arguments. */
export function splitCommand(input: string): { cmd: string; args: string[] } {
  const parts = input.trim().split(/\s+/);
  if (parts.length === 0 || (parts.length === 1 && parts[0] === '')) {
    return { cmd: '', args: [] };
  }
  return { cmd: parts[0].toLowerCase(), args: parts.slice(1) };
}

/** Parse a single integer argument at the given index. */
export function parseIntArg(args: string[], index: number): { value: number } | { error: string } {
  if (index >= args.length) {
    return { error: `Missing argument at position ${index}` };
  }
  const n = Number(args[index]);
  if (!Number.isInteger(n)) {
    return { error: `Invalid number: ${args[index]}` };
  }
  return { value: n };
}

/** Parse all arguments as integers. */
export function parseIntSlice(args: string[]): { values: number[] } | { error: string } {
  const values: number[] = [];
  for (const a of args) {
    const n = Number(a);
    if (!Number.isInteger(n)) {
      return { error: `Invalid number: ${a}` };
    }
    values.push(n);
  }
  return { values };
}

/** Compute Levenshtein distance between two strings. */
function levenshtein(a: string, b: string): number {
  const m = a.length;
  const n = b.length;
  const dp: number[][] = Array.from({ length: m + 1 }, () => Array(n + 1).fill(0));
  for (let i = 0; i <= m; i++) dp[i][0] = i;
  for (let j = 0; j <= n; j++) dp[0][j] = j;
  for (let i = 1; i <= m; i++) {
    for (let j = 1; j <= n; j++) {
      dp[i][j] = a[i - 1] === b[j - 1] ? dp[i - 1][j - 1] : 1 + Math.min(dp[i - 1][j], dp[i][j - 1], dp[i - 1][j - 1]);
    }
  }
  return dp[m][n];
}

/** Suggest the closest command from the list, or null if none is close enough. */
export function suggestCommand(input: string, commands: string[]): string | null {
  if (input === '') return null;
  let best = '';
  let bestDist = Number.POSITIVE_INFINITY;
  for (const cmd of commands) {
    const d = levenshtein(input, cmd);
    if (d < bestDist) {
      bestDist = d;
      best = cmd;
    }
  }
  const maxDist = Math.max(2, Math.floor(input.length / 2));
  return bestDist <= maxDist ? best : null;
}
