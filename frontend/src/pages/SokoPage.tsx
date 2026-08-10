import { withTutorial } from '../components/tutorial/withTutorial';
import type { TutorialStep } from '../types/tutorial';
import { FiveCardStudPageContent } from './FiveCardStudPage';

/**
 * Soko tutorial steps.
 *
 * The deal, the streets and the betting are identical to Five Card Stud, so the
 * steps that differ are the ones about the two extra hand ranks: what a
 * four-card straight and a four-card flush are, and where they sit.
 */
const SOKO_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="fcs-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="fcs-player-hand"]', messageKey: 'tutorial.fourCard', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="fcs-action-buttons"]',
    messageKey: 'tutorial.ranking',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fcs-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/**
 * Renders the Soko (Canadian Stud) page.
 *
 * The whole page is shared with Five Card Stud — same deal, same streets, same
 * betting — and only the showdown ranking differs, which the server has already
 * resolved into `handName` and `handRank` by the time the page sees it. So this
 * passes a game key and the shared content renders it.
 */
export const SokoPage = withTutorial(() => <FiveCardStudPageContent gameKey="soko" />, 'soko', SOKO_TUTORIAL_STEPS);
