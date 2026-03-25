import { useTranslation } from 'react-i18next';
import { jokerpokerApi } from '../api/gameApi';
import { VideoPokerGameContent } from '../components/VideoPokerGameContent';
import { TutorialProvider } from '../providers/TutorialProvider';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';

/** Joker Poker payout table rows. */
const JP_PAYOUT_ROWS = [
  'naturalRoyalFlush5',
  'naturalRoyalFlush',
  'fiveOfAKind',
  'wildRoyalFlush',
  'straightFlush',
  'fourOfAKind',
  'fullHouse',
  'flush',
  'straight',
  'threeOfAKind',
  'twoPair',
  'kingsOrBetter',
];

/** Joker Poker tutorial step definitions. */
const JP_TUTORIAL_STEPS: TutorialStep[] = [
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

/** Joker Poker tutorial configuration. */
const JP_TUTORIAL_CONFIG: TutorialConfig = {
  gameName: 'jokerpoker',
  steps: JP_TUTORIAL_STEPS,
};

/** Renders the Joker Poker (Kings or Better) game page. */
export function JokerPokerPage() {
  const { t: tJp } = useTranslation('jokerpoker');
  return (
    <TutorialProvider config={JP_TUTORIAL_CONFIG} translateMessage={tJp}>
      <VideoPokerGameContent
        gameName="jokerpoker"
        i18nNamespace="jokerpoker"
        apiExec={jokerpokerApi.exec}
        payoutTableRows={JP_PAYOUT_ROWS}
      />
    </TutorialProvider>
  );
}
