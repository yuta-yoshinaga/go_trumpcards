import { describe, expect, it } from 'vitest';
import { loadCuiManualText } from './cuiManualTexts';
import { loadManualText } from './manualTexts';

// These load the real Markdown through `import.meta.glob`, deliberately not a
// mock: the whole point of the change is that the route path now *computes*
// the filename instead of being spelled out in a 219-entry literal, so the
// thing worth testing is that the computed lookup actually resolves. A mocked
// module map would test nothing.
describe('loadManualText', () => {
  it('resolves the BlackJack manual for the root route', async () => {
    await expect(loadManualText('/')).resolves.not.toBe('');
  });

  it('resolves a manual for a named route', async () => {
    await expect(loadManualText('/poker')).resolves.not.toBe('');
  });

  it('returns a different manual per game rather than one shared blob', async () => {
    const [root, poker] = await Promise.all([loadManualText('/'), loadManualText('/poker')]);
    expect(root).not.toBe(poker);
  });

  it('returns an empty string for a route with no manual', async () => {
    await expect(loadManualText('/nosuchgame')).resolves.toBe('');
  });
});

describe('loadCuiManualText', () => {
  it('resolves the CUI manual for the root route', async () => {
    await expect(loadCuiManualText('/')).resolves.not.toBe('');
  });

  it('is a different document from the web manual for the same game', async () => {
    const [web, cui] = await Promise.all([loadManualText('/poker'), loadCuiManualText('/poker')]);
    expect(cui).not.toBe(web);
  });

  it('returns an empty string for a route with no manual', async () => {
    await expect(loadCuiManualText('/nosuchgame')).resolves.toBe('');
  });
});
