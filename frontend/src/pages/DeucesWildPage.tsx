import { useMemo } from 'react';
import { deuceswildApi } from '../api/gameApi';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { VideoPokerGameContent } from '../components/VideoPokerGameContent';
import { VIDEO_POKER_TUTORIAL_STEPS } from '../constants/videoPokerTutorial';
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

/** Renders the Deuces Wild game page. */
export function DeucesWildPage() {
  const cliGameConfig = useMemo(
    () => ({
      parseCommand: parseDeuceswildCommand,
      formatResponse: formatDeuceswildState,
      helpText: DEUCESWILD_HELP,
    }),
    [],
  );
  return (
    <TutorialWrapper gameName="deuceswild" steps={VIDEO_POKER_TUTORIAL_STEPS}>
      <VideoPokerGameContent
        gameName="deuceswild"
        i18nNamespace="deuceswild"
        apiExec={deuceswildApi.exec}
        payoutTableRows={DW_PAYOUT_ROWS}
        gamePath="/deuceswild"
        cliGameConfig={cliGameConfig}
      />
    </TutorialWrapper>
  );
}
