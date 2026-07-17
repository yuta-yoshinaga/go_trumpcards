import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { isWizardLegalPlay, isWizardSpecial, wizardLeadSuit } from './wizardLegal';

/** Build a normal suit card. */
const c = (design: Card['design'], value: number): Card => ({ design, value });
/** Build a Wizard special card (design JOKER, label "Wizard"). */
const wizard = (): Card => ({ design: 'JOKER', value: 1, label: 'Wizard', deck: 'wizard' });
/** Build a Jester special card (design JOKER, label "Jester"). */
const jester = (): Card => ({ design: 'JOKER', value: 1, label: 'Jester', deck: 'wizard' });

describe('isWizardSpecial', () => {
  it('is true for Wizard and Jester cards', () => {
    expect(isWizardSpecial(wizard())).toBe(true);
    expect(isWizardSpecial(jester())).toBe(true);
  });

  it('is false for normal suit cards', () => {
    expect(isWizardSpecial(c('SPADE', 7))).toBe(false);
  });
});

describe('wizardLeadSuit', () => {
  it('returns null for an empty trick', () => {
    expect(wizardLeadSuit([])).toBeNull();
  });

  it('returns null when a Wizard led', () => {
    expect(wizardLeadSuit([wizard(), c('HEART', 5)])).toBeNull();
  });

  it('skips Jesters and uses the first normal suit card', () => {
    expect(wizardLeadSuit([jester(), c('DIAMOND', 9)])).toBe('DIAMOND');
  });

  it('returns null when only Jesters have been played', () => {
    expect(wizardLeadSuit([jester(), jester()])).toBeNull();
  });

  it('returns the first normal card design', () => {
    expect(wizardLeadSuit([c('CLOVER', 3), c('HEART', 2)])).toBe('CLOVER');
  });
});

describe('isWizardLegalPlay', () => {
  it('allows any card when leading (empty trick)', () => {
    const hand = [c('SPADE', 4), c('HEART', 9)];
    expect(isWizardLegalPlay(c('SPADE', 4), [], hand)).toBe(true);
    expect(isWizardLegalPlay(c('HEART', 9), [], hand)).toBe(true);
  });

  it('always allows a Wizard or Jester card', () => {
    const trick = [c('HEART', 5)];
    const hand = [c('HEART', 2), wizard(), jester()];
    expect(isWizardLegalPlay(wizard(), trick, hand)).toBe(true);
    expect(isWizardLegalPlay(jester(), trick, hand)).toBe(true);
  });

  it('requires following the led suit when the player holds it', () => {
    const trick = [c('HEART', 5)];
    const hand = [c('HEART', 2), c('SPADE', 9)];
    expect(isWizardLegalPlay(c('HEART', 2), trick, hand)).toBe(true); // follows
    expect(isWizardLegalPlay(c('SPADE', 9), trick, hand)).toBe(false); // must follow HEART
  });

  it('allows any card when void in the led suit', () => {
    const trick = [c('HEART', 5)];
    const hand = [c('SPADE', 9), c('DIAMOND', 3)];
    expect(isWizardLegalPlay(c('SPADE', 9), trick, hand)).toBe(true);
    expect(isWizardLegalPlay(c('DIAMOND', 3), trick, hand)).toBe(true);
  });

  it('allows any card when a Wizard led (no led suit)', () => {
    const trick = [wizard()];
    const hand = [c('HEART', 2), c('SPADE', 9)];
    expect(isWizardLegalPlay(c('SPADE', 9), trick, hand)).toBe(true);
    expect(isWizardLegalPlay(c('HEART', 2), trick, hand)).toBe(true);
  });

  it('treats an all-Jester trick as having no led suit', () => {
    const trick = [jester()];
    const hand = [c('HEART', 2), c('SPADE', 9)];
    expect(isWizardLegalPlay(c('SPADE', 9), trick, hand)).toBe(true);
  });
});
