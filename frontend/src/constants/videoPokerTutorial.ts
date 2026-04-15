import type { TutorialStep } from '../types/tutorial';

/** Shared tutorial step definitions for Video Poker variants (VP, Deuces Wild, Joker Poker). */
export const VIDEO_POKER_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="vp-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="vp-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="vp-draw-button"]',
    messageKey: 'tutorial.drawButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="vp-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];
