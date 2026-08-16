import i18n from '../../i18n';
import type { HintResult } from '../../types/hint';

/**
 * Commands that ask for a hint, matching what the parser modules accept.
 *
 * `localCommand` runs *before* `parseCommand`, so this must stay narrow: any
 * input it claims never reaches the game's own parser.
 */
const HINT_COMMANDS = new Set(['hint', 'h']);

/** Reports whether a CLI input is asking for a hint. */
export function isHintCommand(input: string): boolean {
  return HINT_COMMANDS.has(input.trim().toLowerCase());
}

/**
 * Render a {@link HintResult} as a line for the Web CLI terminal.
 *
 * These games have no `hint` action on their backend -- `useGameHint` computes
 * the advice client-side from state the page already holds -- so the CLI answers
 * locally instead of round-tripping to an endpoint that would reject it (#5793).
 *
 * The hint logic itself is never duplicated here: this only formats what
 * `useGameHint` already decided.
 */
export function hintCliText(hint: HintResult | null): string {
  if (!hint) return i18n.t('cli.hintNone');

  // Pass reasonParams. Several pages render the GUI tooltip with a bare
  // t(hint.reason), which silently drops the interpolation values and leaves
  // "{{zone}}" or a generic sentence on screen (#4885 is the same shape).
  const parts = [i18n.t(hint.reason, hint.reasonParams ?? {})];

  if (hint.targetPos !== undefined) {
    parts.push(i18n.t('cli.hintTarget', { pos: hint.targetPos }));
  } else if (hint.targetIndices?.length) {
    parts.push(i18n.t('cli.hintTargets', { positions: hint.targetIndices.join(', ') }));
  }

  parts.push(i18n.t(`cli.hintConfidence.${hint.confidence}`));
  return parts.join(' ');
}

/**
 * Builds the `localCommand` a page hands to {@link useCliGame}.
 *
 * Every page needs the identical three-token lambda, and 73 copies of it are 73
 * lines no test executes -- the page tests never enter CLI mode. Returning it
 * from one tested factory keeps the behaviour covered and the call sites to a
 * single argument.
 *
 * Returns `null` for anything that is not a hint request, which is the contract
 * `localCommand` uses to fall through to the game's own parser.
 */
export function hintLocalCommand(hint: HintResult | null): (input: string) => string | null {
  return (input: string) => (isHintCommand(input) ? hintCliText(hint) : null);
}
