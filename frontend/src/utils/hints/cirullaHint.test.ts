import { describe, expect, it } from 'vitest';
import { makeCirullaState } from '../../test/stateFactories';
import { getCirullaHint } from './cirullaHint';

describe('getCirullaHint', () => {
  it.each(['capture', 'sweep', 'lay_off'])('maps the %s reason into a hint key', (hintReason) => {
    expect(getCirullaHint(makeCirullaState({ hintHandIdx: 1, hintReason }))).toEqual({
      targetAction: 'play',
      reason: `hint.${hintReason}`,
      confidence: 'moderate',
    });
  });

  // **手札 0 は正当な助言。** -1 との区別を落とすと先頭の札を勧められない。
  it('keeps a suggestion for hand index 0', () => {
    expect(getCirullaHint(makeCirullaState({ hintHandIdx: 0, hintReason: 'capture' }))).not.toBeNull();
  });

  it('returns null when the server sent no suggestion', () => {
    expect(getCirullaHint(makeCirullaState())).toBeNull();
    expect(getCirullaHint(makeCirullaState({ hintHandIdx: -1, hintReason: 'capture' }))).toBeNull();
    expect(getCirullaHint(makeCirullaState({ hintHandIdx: 1, hintReason: '' }))).toBeNull();
  });

  it('returns null for the explicit no-suggestion reason', () => {
    expect(getCirullaHint(makeCirullaState({ hintHandIdx: 1, hintReason: 'none' }))).toBeNull();
  });
});
