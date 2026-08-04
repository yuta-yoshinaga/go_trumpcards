/** The parts of `MaoResponse` the hint text is built from. */
export interface MaoRuleHintSource {
  ruleHint: string;
  ruleHintCode?: string;
}

/**
 * The half-hint to show, in the browser's language.
 *
 * The server's i18n language is process-wide (`i18n.SetLang` is called once at
 * startup), so `ruleHint` always arrives in the server's language regardless of
 * what the browser is set to. `ruleHintCode` carries the key instead, and this
 * translates it here (#4917). `ruleHint` stays as the fallback for a response
 * from a server old enough not to send the code.
 * @param state - The Mao response.
 * @param t - The `mao` namespace translator.
 * @returns The hint text, or the server's string when the code cannot be translated.
 */
export function ruleHintText(state: MaoRuleHintSource, t: (key: string) => string): string {
  if (!state.ruleHintCode) return state.ruleHint;
  const key = `ruleHint.${state.ruleHintCode}`;
  const translated = t(key);
  return translated === key ? state.ruleHint : translated;
}
