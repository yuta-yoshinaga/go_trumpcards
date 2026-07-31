import { useEffect, useMemo, useState } from 'react';
import { bostonApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BostonBidOption, BostonResponse } from '../types/card';
import { BostonPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { BOSTON_HELP, parseBostonCommand } from '../utils/cli/commands/bostonCommands';
import { formatBostonState } from '../utils/cli/formatters/bostonFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Suit glyphs by design value (1=Spade … 4=Diamond). */
const SUIT_GLYPHS: Readonly<Record<number, string>> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Bid kinds as sent on the wire (sync: `BostonBidKind`). */
const KIND_TRICKS = 1;
const KIND_MISERE = 2;

/** Boston tutorial step definitions. */
const BOSTON_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="boston-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="boston-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="boston-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="boston-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="boston-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const BOSTON_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BostonPhase.BID]: 'bid',
  [BostonPhase.CALL_PARTNER]: 'callPartner',
  [BostonPhase.PLAY]: 'play',
  [BostonPhase.HAND_END]: 'handEnd',
  [BostonPhase.GAME_END]: 'gameEnd',
};

/** Renders the Boston game page: the 18th-century Whist derivative. */
export const BostonPage = withTutorial(BostonPageContent, 'boston', BOSTON_TUTORIAL_STEPS);

/** Inner content of the Boston page, wrapped by TutorialProvider. */
function BostonPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('boston');
  const { state, loading, error, exec, retry } = useGameApi(bostonApi.exec);

  const [bidLevel, setBidLevel] = useState<number | null>(null);
  const [bidSuit, setBidSuit] = useState(1);
  const [selected, setSelected] = useState<number | null>(null);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('boston');
  const cliConfig: CliGameConfig<BostonResponse, Parameters<typeof bostonApi.exec>> = useMemo(
    () => ({
      gameName: 'boston',
      parseCommand: parseBostonCommand,
      formatResponse: formatBostonState,
      helpText: BOSTON_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('boston', BOSTON_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="boston" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 13 }} />;

  const human = state.players.find((p) => p.isHuman);
  const isBid = state.phase === BostonPhase.BID;
  const isCallPartner = state.phase === BostonPhase.CALL_PARTNER;
  const isPlay = state.phase === BostonPhase.PLAY;
  const isHandEnd = state.phase === BostonPhase.HAND_END;
  const isGameEnd = state.phase === BostonPhase.GAME_END || state.gameEndFlag;
  const humanWon = isGameEnd && state.winnerIdx === 0;
  const isHumanBid = isBid && state.bidPlayerIdx === 0 && !isGameEnd;
  const isHumanCallPartner = isCallPartner && state.declarerIdx === 0 && !isGameEnd;
  const isHumanPlay = isPlay && state.currentPlayerIdx === 0 && !isGameEnd;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));
  const suitGlyph = (suit: number): string => SUIT_GLYPHS[suit] ?? t('noTrump');
  const bidLabel = (name: string): string => t(`bid.${name}`);

  /** What a bid asks of you, in words — the three kinds differ in win condition. */
  const kindLabel = (o: BostonBidOption): string => {
    if (o.kind === KIND_TRICKS) return t('kindTricks', { n: o.tricks });
    if (o.kind === KIND_MISERE) return t('kindMisere');
    return t('kindPiccolissimo');
  };

  // **立っている宣言より上しか選べない。**
  const selectable = state.bidOptions.filter((o) => !state.highBid || o.level > state.highBid.level);
  const chosen = state.bidOptions.find((o) => o.level === bidLevel) ?? null;

  const canPlay = (i: number) => state.validPlays.includes(i);

  const handleBid = () => {
    if (bidLevel === null) return;
    exec('bid', { level: bidLevel, suit: chosen?.needsTrump ? bidSuit : undefined });
    setBidLevel(null);
  };

  const handlePlay = () => {
    if (selected === null) return;
    exec('play', { cardIndex: selected });
    setSelected(null);
  };

  const handleManualReset = () => {
    hideActionLog();
    setSelected(null);
    setBidLevel(null);
    exec('reset');
  };

  return (
    <GamePageShell
      title={tc('nav.boston')}
      gameThemeBg={gameTheme.boston.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanBid || isHumanCallPartner || isHumanPlay}
      gamePath="/boston"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="boston-info">
              <span className="mr-4">{t('hand', { n: state.handNumber, target: state.targetHands })}</span>
              {state.highBid && (
                <span className="mr-4" data-testid="boston-contract">
                  {t('contract')}: {bidLabel(state.highBid.name)} {suitGlyph(state.trumpSuit)}
                </span>
              )}
            </div>

            {/* The ladder — the one thing the ordering of this game turns on. */}
            <div className="mb-2 p-2 rounded bg-black/20 text-xs" data-testid="boston-ladder">
              <div className="mb-1 text-ds-text-primary">{t('ladderTitle')}</div>
              <ol className="flex flex-col gap-0.5">
                {state.bidOptions.map((o) => (
                  <li
                    key={o.level}
                    className={o.kind === KIND_TRICKS ? 'text-ds-text-muted' : 'text-ds-warning font-semibold'}
                  >
                    {o.level}. {bidLabel(o.name)} — {kindLabel(o)}
                    {o.exposed && ` [${t('exposedTag')}]`}
                    {o.canCallPartner && ` [${t('partnerTag')}]`} ({t('payout', { n: o.payout })})
                  </li>
                ))}
              </ol>
              <div className="mt-1 text-ds-text-muted">{t('ladderNote')}</div>
            </div>

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="boston-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p) => (
                <div
                  key={`player-${p.id}`}
                  data-testid="boston-player"
                  className={`text-sm py-0.5 flex items-center gap-2 ${
                    p.isCurrentTurn && !isGameEnd ? 'text-ds-warning' : 'text-ds-text-muted'
                  } ${p.isHuman ? 'font-semibold' : ''}`}
                >
                  <span>{playerLabel(p.id, p.isHuman)}</span>
                  {p.isDealer && <span className="text-ds-accent">[{t('dealer')}]</span>}
                  {p.isDeclarer && <span className="text-ds-success">[{t('declarer')}]</span>}
                  {p.isPartner && <span className="text-ds-success">[{t('partner')}]</span>}
                  <span>{t('tricksWon', { n: p.tricksWon })}</span>
                  <span>{t('chips', { n: p.chips })}</span>
                  {!p.isHuman && p.cards.length === 0 && <span>{t('hiddenHand', { count: p.cardCount })}</span>}
                </div>
              ))}
            </div>

            {/* Trick */}
            {state.trick.length > 0 && (
              <div className="mb-2 flex items-center gap-2" data-testid="boston-trick">
                <span className="text-ds-text-muted text-sm">{t('trick')}</span>
                {state.trick.map((c, i) => (
                  <CardImage key={`trick-${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />
                ))}
              </div>
            )}

            {/* Settlement */}
            {(isHandEnd || isGameEnd) && (
              <div
                className={`mb-2 p-2 rounded text-sm ${badgeWarningColors}`}
                data-testid="boston-settlement"
                role="status"
                aria-live="polite"
              >
                <div className="mb-1">{t('settlementTitle')}</div>
                <div>
                  {state.bidMade
                    ? t('madeLine', { tricks: state.declarerTricks })
                    : t('failedLine', { tricks: state.declarerTricks })}
                </div>
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.boston.footer} px-4 py-2.5`}>
            <div className="mb-2" data-tutorial="boston-hand">
              <div className="text-ds-text-muted text-xs mb-1">{t('yourHand')}</div>
              <div className="flex flex-wrap gap-1">
                {(human?.cards ?? []).map((c, i) => (
                  <button
                    key={`hand-${c.design}-${c.value}-${i}`}
                    type="button"
                    data-hint-action="play"
                    onClick={() => setSelected(i)}
                    disabled={loading || (isPlay && !canPlay(i))}
                    className={`rounded ${selected === i ? 'ring-2 ring-ds-accent' : ''} ${
                      isPlay && !canPlay(i) ? 'opacity-40' : ''
                    }`}
                  >
                    <CardImage card={c} width={cardWidth} />
                  </button>
                ))}
              </div>
            </div>

            <ErrorAlert message={error} onRetry={retry} />

            {isHumanBid && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="boston-bid-notice">
                {t('bidNotice')}
              </div>
            )}
            {isHumanCallPartner && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="boston-partner-notice">
                {t('partnerNotice')}
              </div>
            )}
            {isHumanPlay && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="boston-play-notice">
                {t('playNotice')}
              </div>
            )}

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="boston-actions">
              {isHumanBid && (
                <>
                  <label className="text-ds-text-muted text-xs flex items-center gap-1" htmlFor="boston-bid-select">
                    {t('contract')}
                    <select
                      id="boston-bid-select"
                      className="bg-black/30 text-ds-text-primary rounded px-1 min-h-[44px]"
                      value={bidLevel ?? ''}
                      onChange={(e) => setBidLevel(e.target.value === '' ? null : Number(e.target.value))}
                    >
                      <option value="">—</option>
                      {selectable.map((o) => (
                        <option key={o.level} value={o.level}>
                          {o.level}. {bidLabel(o.name)}
                        </option>
                      ))}
                    </select>
                  </label>

                  {/* **切札が要るのはトリック宣言だけ。**ミゼールでは出さない。 */}
                  {chosen?.needsTrump && (
                    <label className="text-ds-text-muted text-xs flex items-center gap-1" htmlFor="boston-suit-select">
                      {t('suitLabel')}
                      <select
                        id="boston-suit-select"
                        className="bg-black/30 text-ds-text-primary rounded px-1 min-h-[44px]"
                        value={bidSuit}
                        onChange={(e) => setBidSuit(Number(e.target.value))}
                      >
                        {[1, 2, 3, 4].map((s) => (
                          <option key={s} value={s}>
                            {SUIT_GLYPHS[s]}
                          </option>
                        ))}
                      </select>
                    </label>
                  )}

                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleBid}
                    disabled={loading || bidLevel === null}
                  >
                    {t('bidButton')}
                  </button>
                  <button type="button" className={btnWarning} onClick={() => exec('pass')} disabled={loading}>
                    {t('passButton')}
                  </button>
                </>
              )}

              {isHumanCallPartner && (
                <>
                  {state.players
                    .filter((p) => p.id !== state.declarerIdx)
                    .map((p) => (
                      <button
                        key={`partner-${p.id}`}
                        type="button"
                        className={btnPrimary}
                        onClick={() => exec('callpartner', { partner: p.id })}
                        disabled={loading}
                      >
                        {t('callButton', { name: playerLabel(p.id, p.isHuman) })}
                      </button>
                    ))}
                  <button
                    type="button"
                    className={btnWarning}
                    onClick={() => exec('callpartner', { partner: -1 })}
                    disabled={loading}
                  >
                    {t('aloneButton')}
                  </button>
                </>
              )}

              {isHumanPlay && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handlePlay}
                  disabled={loading || selected === null}
                >
                  {t('playButton')}
                </button>
              )}

              {isHandEnd && !isGameEnd && (
                <button type="button" className={btnSuccess} onClick={() => exec('next')} disabled={loading}>
                  {t('nextHand')}
                </button>
              )}

              {isGameEnd && (
                <span className="text-ds-text-primary text-sm font-semibold mr-1">
                  {humanWon ? t('win') : t('lose', { name: playerLabel(state.winnerIdx, state.winnerIdx === 0) })}
                </span>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="boston-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
