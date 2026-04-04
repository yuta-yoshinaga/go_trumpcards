import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { deuceswildApi } from '../api/gameApi';
import { VideoPokerGameContent } from '../components/VideoPokerGameContent';
import { TutorialProvider } from '../providers/TutorialProvider';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';
import { DEUCESWILD_HELP, parseDeuceswildCommand } from '../utils/cli/commands/deuceswildCommands';
import { formatDeuceswildState } from '../utils/cli/formatters/deuceswildFormatter';

/** Deuces Wild payout table rows. */
const DW_PAYOUT_ROWS = [
  'naturalRoyalFlush5',
  'naturalRoyalFlush',
  'fourDeuces',
  'wildRoyalFlush',
  'fiveOfAKind',
  'straightFlush',
  'fourOfAKind',
  'fullHouse',
  'flush',
  'straight',
  'threeOfAKind',
];

/** Deuces Wild tutorial step definitions. */
const DW_TUTORIAL_STEPS: TutorialStep[] = [
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

/** Deuces Wild tutorial configuration. */
const DW_TUTORIAL_CONFIG: TutorialConfig = {
  gameName: 'deuceswild',
  steps: DW_TUTORIAL_STEPS,
};

/** Renders the Deuces Wild game page. */
export function DeucesWildPage() {
  const { t: tDw } = useTranslation('deuceswild');
  const cliGameConfig = useMemo(
    () => ({
      parseCommand: parseDeuceswildCommand,
      formatResponse: formatDeuceswildState,
      helpText: DEUCESWILD_HELP,
    }),
    [],
  );
  return (
    <TutorialProvider config={DW_TUTORIAL_CONFIG} translateMessage={tDw}>
      <VideoPokerGameContent
        gameName="deuceswild"
        i18nNamespace="deuceswild"
        apiExec={deuceswildApi.exec}
        payoutTableRows={DW_PAYOUT_ROWS}
        gamePath="/deuceswild"
        cliGameConfig={cliGameConfig}
      />
    </TutorialProvider>
  );
}
