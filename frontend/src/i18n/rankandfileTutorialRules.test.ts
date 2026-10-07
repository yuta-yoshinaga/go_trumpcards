import { describe, expect, it } from 'vitest';
import en from './locales/en/rankandfile.json';
import ja from './locales/ja/rankandfile.json';

describe('rank and file tutorial describes tableau moves', () => {
  it('allows single cards and movable alternating-color descending sequences', () => {
    expect(ja.tutorial.tableau).toContain('色違いで1つ上のランクの札の上に重ねます');
    expect(ja.tutorial.tableau).toContain('1枚でも');
    expect(ja.tutorial.tableau).toContain('色違いの降順');
    expect(ja.tutorial.tableau).toContain('まとまり');
    expect(en.tutorial.tableau).toContain('a single face-up card');
    expect(en.tutorial.tableau).toContain('alternating-color descending sequence');
    expect(en.tutorial.tableau).toContain('Build down in alternating colors');
  });

  it('keeps the rule that any card can go in an empty column', () => {
    expect(ja.tutorial.tableau).toContain('空の列にはどのカードでも置けます');
    expect(en.tutorial.tableau).toContain('Any card can fill an empty column.');
  });
});
