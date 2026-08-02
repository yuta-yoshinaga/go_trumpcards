/**
 * True when this response is an answer to an explicit `hint` command.
 *
 * Since #4483, `Output()` also carries the hint so the board tooltip can read
 * `state.hint` (it was permanently undefined before). But `Output()` runs on
 * every command, so a CLI formatter that keys off `state.hint` alone would
 * print `HINT:` after every single move, unasked.
 *
 * Only the `hint` command's own response sets a `hintAvailable` message code,
 * so that is what separates "the player asked" from "the tooltip needs data".
 * Gating on it costs nothing: no extra request, and no extra KV write, which
 * matters because every request writes the session back and the free tier
 * allows 1,000 writes a day (ADR-0028).
 *
 * This lived in `cli/formatterBase.ts` while only the CLI used it. The Web GUI
 * needs the same distinction and did not have it, so 42 game pages showed the
 * hint banner on every turn to players who never asked (#4605); it moved here
 * so both sides read one definition.
 */
export function isRequestedHint(state: { messageCode?: string }): boolean {
  // **2 つある。**ソリティア側は `<game>.hintAvailable` を使うが、
  // トリックテイキング側ではそのキーが既に**ラベル**（「ヒント」）として
  // 埋まっていて、メッセージコードに流用すると意味が二重になる (#4483)。
  // そちらは `<game>.hintRequested` を出す。
  const code = state.messageCode;
  return code?.endsWith('.hintAvailable') === true || code?.endsWith('.hintRequested') === true;
}
