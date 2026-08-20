import { describe, expect, it } from 'vitest';
import i18n from '../../i18n';
import type { HintResult } from '../../types/hint';
import { hintCliText, hintLocalCommand, isHintCommand } from './hintText';

const base: HintResult = {
  targetAction: 'play',
  reason: 'hintText.test.plain',
  confidence: 'strong',
};

// Register fixture strings on the live instance so the test exercises the real
// i18n path (interpolation included) rather than a stub that cannot drop params.
i18n.addResource('ja', 'common', 'hintText.test.plain', 'いちばん強い札を出す');
i18n.addResource('ja', 'common', 'hintText.test.withParams', '{{zone}} の {{n}} 番へ');
i18n.addResource('en', 'common', 'hintText.test.plain', 'Play your strongest card');
i18n.addResource('en', 'common', 'hintText.test.withParams', 'to {{zone}} slot {{n}}');

describe('isHintCommand', () => {
  it.each(['hint', 'h', 'HINT', ' Hint '])('accepts %j', (input) => {
    expect(isHintCommand(input)).toBe(true);
  });

  it.each(['hints', 'help', '', 'p 3', 'hi'])('rejects %j', (input) => {
    expect(isHintCommand(input)).toBe(false);
  });
});

describe('hintCliText', () => {
  it('renders the reason', () => {
    expect(hintCliText(base)).toContain('いちばん強い札を出す');
  });

  it('interpolates reasonParams', () => {
    // The GUI calls t(hint.reason) without params on several pages, which drops
    // the values entirely (same shape as #4885). The CLI must not repeat that.
    const out = hintCliText({
      ...base,
      reason: 'hintText.test.withParams',
      reasonParams: { zone: '組札', n: 2 },
    });
    expect(out).toContain('組札');
    expect(out).toContain('2');
    expect(out).not.toContain('{{');
  });

  it('says so explicitly when there is no hint', () => {
    const out = hintCliText(null);
    expect(out.length).toBeGreaterThan(0);
    expect(out).not.toBe('');
  });

  it('reports the target position when the hint knows one', () => {
    const out = hintCliText({ ...base, targetPos: 4 });
    expect(out).toContain('4');
  });

  it('omits position wording when the hint has none', () => {
    const withPos = hintCliText({ ...base, targetPos: 4 });
    const without = hintCliText(base);
    expect(without.length).toBeLessThan(withPos.length);
  });

  it('reports several target positions', () => {
    const out = hintCliText({ ...base, targetIndices: [0, 2, 5] });
    for (const n of ['0', '2', '5']) expect(out).toContain(n);
  });

  it('distinguishes confidence levels', () => {
    const strong = hintCliText({ ...base, confidence: 'strong' });
    const moderate = hintCliText({ ...base, confidence: 'moderate' });
    expect(strong).not.toBe(moderate);
  });

  it('follows a language switch instead of caching the first render', async () => {
    const prev = i18n.language;
    await i18n.changeLanguage('en');
    const en = hintCliText({
      ...base,
      reason: 'hintText.test.withParams',
      reasonParams: { zone: 'foundation', n: 2 },
    });
    expect(en).toContain('foundation');
    await i18n.changeLanguage('ja');
    const ja = hintCliText({
      ...base,
      reason: 'hintText.test.withParams',
      reasonParams: { zone: '組札', n: 2 },
    });
    expect(ja).toContain('組札');
    expect(ja).not.toContain('foundation');
    await i18n.changeLanguage(prev);
  });
});

describe('hintLocalCommand', () => {
  it('answers a hint request', () => {
    const cmd = hintLocalCommand(base);
    expect(cmd('hint')).toContain('いちばん強い札を出す');
    expect(cmd('h')).toContain('いちばん強い札を出す');
  });

  it('falls through for anything else, so the game parser still sees it', () => {
    const cmd = hintLocalCommand(base);
    for (const other of ['p 3', 'reset', 'help', '']) expect(cmd(other)).toBeNull();
  });

  it('still answers when there is no hint, rather than falling through silently', () => {
    // Returning null here would hand "hint" to the parser, which rejects it --
    // the player would see "Unknown command" instead of "no hint available".
    expect(hintLocalCommand(null)('hint')).not.toBeNull();
  });
});
