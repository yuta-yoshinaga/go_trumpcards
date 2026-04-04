import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { videopokerApi } from '../api/gameApi';
import { VideoPokerGameContent } from '../components/VideoPokerGameContent';
import { TutorialProvider } from '../providers/TutorialProvider';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';
import { parseVideopokerCommand, VIDEOPOKER_HELP } from '../utils/cli/commands/videopokerCommands';
import { formatVideopokerState } from '../utils/cli/formatters/videopokerFormatter';

/** Jacks or Better payout table rows. */
const JOB_PAYOUT_ROWS = [
  'royalFlush5',
  'royalFlush',
  'straightFlush',
  'fourOfAKind',
  'fullHouse',
  'flush',
  'straight',
  'threeOfAKind',
  'twoPair',
  'jacksOrBetter',
];

/** Video Poker tutorial step definitions. */
const VP_TUTORIAL_STEPS: TutorialStep[] = [
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

/** Video Poker tutorial configuration. */
const VP_TUTORIAL_CONFIG: TutorialConfig = {
  gameName: 'videopoker',
  steps: VP_TUTORIAL_STEPS,
};

/** Renders the Video Poker (Jacks or Better) game page. */
export function VideoPokerPage() {
  const { t: tVp } = useTranslation('videopoker');
  const cliGameConfig = useMemo(
    () => ({
      parseCommand: parseVideopokerCommand,
      formatResponse: formatVideopokerState,
      helpText: VIDEOPOKER_HELP,
    }),
    [],
  );
  return (
    <TutorialProvider config={VP_TUTORIAL_CONFIG} translateMessage={tVp}>
      <VideoPokerGameContent
        gameName="videopoker"
        i18nNamespace="videopoker"
        apiExec={videopokerApi.exec}
        payoutTableRows={JOB_PAYOUT_ROWS}
        gamePath="/videopoker"
        cliGameConfig={cliGameConfig}
      />
    </TutorialProvider>
  );
}
