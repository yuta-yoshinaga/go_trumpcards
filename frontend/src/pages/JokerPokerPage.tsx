import { useMemo } from 'react';
import { jokerpokerApi } from '../api/gameApi';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { VideoPokerGameContent } from '../components/VideoPokerGameContent';
import { VIDEO_POKER_TUTORIAL_STEPS } from '../constants/videoPokerTutorial';
import { JOKERPOKER_HELP, parseJokerpokerCommand } from '../utils/cli/commands/jokerpokerCommands';
import { formatJokerpokerState } from '../utils/cli/formatters/jokerpokerFormatter';

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

/** Renders the Joker Poker (Kings or Better) game page. */
export function JokerPokerPage() {
  const cliGameConfig = useMemo(
    () => ({
      parseCommand: parseJokerpokerCommand,
      formatResponse: formatJokerpokerState,
      helpText: JOKERPOKER_HELP,
    }),
    [],
  );
  return (
    <TutorialWrapper gameName="jokerpoker" steps={VIDEO_POKER_TUTORIAL_STEPS}>
      <VideoPokerGameContent
        gameName="jokerpoker"
        i18nNamespace="jokerpoker"
        apiExec={jokerpokerApi.exec}
        payoutTableRows={JP_PAYOUT_ROWS}
        gamePath="/jokerpoker"
        cliGameConfig={cliGameConfig}
      />
    </TutorialWrapper>
  );
}
