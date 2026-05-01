import { deuceswildApi } from '../api/gameApi';
import { withTutorial } from '../components/tutorial/withTutorial';
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

const CLI_GAME_CONFIG = {
  parseCommand: parseDeuceswildCommand,
  formatResponse: formatDeuceswildState,
  helpText: DEUCESWILD_HELP,
} as const;

function DeucesWildPageContent() {
  return (
    <VideoPokerGameContent
      gameName="deuceswild"
      i18nNamespace="deuceswild"
      apiExec={deuceswildApi.exec}
      payoutTableRows={DW_PAYOUT_ROWS}
      gamePath="/deuceswild"
      cliGameConfig={CLI_GAME_CONFIG}
    />
  );
}

/** Renders the Deuces Wild game page. */
export const DeucesWildPage = withTutorial(DeucesWildPageContent, 'deuceswild', VIDEO_POKER_TUTORIAL_STEPS);
