import type { ohHellApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { MobileHandGrid } from '../components/MobileHandGrid';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { ScrollFadeHint } from '../components/ScrollFadeHint';
import { TrickDisplay } from '../components/TrickDisplay';
import { useBreakpoint, useCardDimensions } from '../hooks/useCardDimensions';
import {
  CPU_DIFFICULTY_OPTIONS,
  MAX_HAND_SIZE_OPTIONS,
  ROUND_DIRECTION_OPTIONS,
  SCORING_VARIANT_OPTIONS,
  useOhHellGame,
} from '../hooks/useOhHellGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeErrorColors, badgeInfoColors, badgeSuccessColors, badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { scoreTableClass } from '../styles/scoreTableStyles';
import type { OhHellResponse } from '../types/card';
import { OhHellPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { OHHELL_HELP, parseOhhellCommand } from '../utils/cli/commands/ohhellCommands';
import { formatOhhellState } from '../utils/cli/formatters/ohhellFormatter';
import { ohHellBidSummary } from '../utils/ohHellBid';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Oh Hell tutorial step definitions. */
const OH_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="oh-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oh-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oh-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oh-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oh-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oh-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Pick the badge color tokens for the bid-progress chip: green on target, yellow overshot, red unreachable. */
function progressChipColors(bid: number, won: number, remainingTricks: number): string {
  if (won === bid) return badgeSuccessColors;
  if (won > bid) return badgeWarningColors;
  return bid - won > remainingTricks ? badgeErrorColors : badgeInfoColors;
}

const OH_HELL_PHASE_KEYS: Readonly<Record<number, string>> = {
  [OhHellPhase.BID]: 'bid',
  [OhHellPhase.PLAY]: 'play',
  [OhHellPhase.TRICK_END]: 'trickEnd',
  [OhHellPhase.ROUND_END]: 'roundEnd',
  [OhHellPhase.GAME_END]: 'gameEnd',
};

type OhHellGame = ReturnType<typeof useOhHellGame>;
type OhHellView = ReturnType<typeof useOhHellView>;
type OhHellContext = GamePageContext<OhHellResponse, OhHellGame, OhHellView>;

function useOhHellView() {
  const isLargeDesktop = useBreakpoint() === 'largeDesktop';
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('ohhell', OH_HELL_PHASE_KEYS);
  return { isLargeDesktop, cardWidth, isMobile, phaseNames };
}

function OhHellMessageBox({ state }: { state: OhHellResponse }) {
  return <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />;
}

function HeaderExtra({ state, t }: OhHellContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPhase = state.phase === OhHellPhase.BID;
  const isRoundEnd = state.phase === OhHellPhase.ROUND_END;
  const isGameEnd = state.phase === OhHellPhase.GAME_END || state.gameEndFlag;

  // A card already played into the unresolved trick still counts as a winnable trick,
  // even though cardCount has been decremented for it.
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const humanInCurrentTrick = state.currentTrick.some((tc) => tc.playerIdx === humanIdx);
  const showProgressChip =
    !isBidPhase && !isRoundEnd && !isGameEnd && humanPlayer !== undefined && humanPlayer.bid >= 0;
  const remainingHumanTricks = humanPlayer ? humanPlayer.cardCount + (humanInCurrentTrick ? 1 : 0) : 0;
  const bidUnreachable =
    humanPlayer !== undefined &&
    (humanPlayer.trickCount > humanPlayer.bid || humanPlayer.bid - humanPlayer.trickCount > remainingHumanTricks);
  // During bidding, summarize the table's placed bids vs the hand size.
  const bidSummary = isBidPhase
    ? ohHellBidSummary(
        state.players.filter((p) => p.bid >= 0).map((p) => p.bid),
        state.handSize,
      )
    : null;

  return (
    <>
      {showProgressChip && (
        <span
          data-testid="bid-progress-chip"
          className={`rounded-full px-2.5 py-1 text-xs font-medium ${progressChipColors(
            humanPlayer.bid,
            humanPlayer.trickCount,
            remainingHumanTricks,
          )}`}
        >
          {t('bidProgress', {
            bid: humanPlayer.bid,
            won: humanPlayer.trickCount,
            remaining: remainingHumanTricks,
          })}
          {bidUnreachable && <span className="ml-1">({t('bidUnreachable')})</span>}
        </span>
      )}
      {bidSummary && (
        <span
          data-testid="bid-total-chip"
          className={`rounded-full px-2.5 py-1 text-xs font-medium ${
            bidSummary.kind === 'over'
              ? badgeWarningColors
              : bidSummary.kind === 'exact'
                ? badgeSuccessColors
                : badgeInfoColors
          }`}
        >
          {t('bidTotal', { total: bidSummary.total, handSize: state.handSize })} {t(`bidOverUnder.${bidSummary.kind}`)}
        </span>
      )}
    </>
  );
}

function Board({ state, t, tc, view }: OhHellContext) {
  const { isLargeDesktop, cardWidth, isMobile } = view;
  const isBidPhase = state.phase === OhHellPhase.BID;
  const isTrickEnd = state.phase === OhHellPhase.TRICK_END;
  const isRoundEnd = state.phase === OhHellPhase.ROUND_END;
  const isGameEnd = state.phase === OhHellPhase.GAME_END || state.gameEndFlag;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;

  const dealerName = playerName(
    state.players[state.dealerIdx]?.id ?? state.dealerIdx,
    state.players[state.dealerIdx]?.isHuman ?? false,
  );

  return (
    <>
      {/* Round/Trick/Trump info */}
      <div
        className="text-ds-text-primary text-center mb-2 flex flex-wrap justify-center gap-x-4 gap-y-1"
        data-testid="oh-status-row"
      >
        <span className="whitespace-nowrap">{t('round', { n: state.roundNumber, total: state.totalRounds })}</span>
        <span className="whitespace-nowrap">{t('trick', { n: state.trickNumber })}</span>
        <span className="whitespace-nowrap">{t('handSize', { n: state.handSize })}</span>
      </div>
      <div className="text-ds-text-primary text-center mb-2 flex flex-wrap justify-center gap-x-4 gap-y-1">
        <span className="whitespace-nowrap">
          {t('trump', { suit: state.trumpSuit > 0 ? t(`suitName.${state.trumpSuit}`) : t('noTrump') })}
        </span>
        <span>{t('dealer', { name: dealerName })}</span>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: game play area */}
        <div>
          {/* Bid phase instruction */}
          {isHumanBidTurn && (
            <div className="text-ds-warning text-center mb-2" data-tutorial="oh-bid-controls">
              <div>{t('bidPhase', { max: state.handSize })}</div>
              {state.restrictedBid >= 0 && (
                <div className="text-ds-warning text-sm">{t('restrictedBid', { n: state.restrictedBid })}</div>
              )}
            </div>
          )}

          {/* Trump card display */}
          {state.trumpCard && (
            <div className="my-2 flex justify-center">
              <div className="text-center">
                <AnimatedCard card={state.trumpCard} width={cardWidth} />
              </div>
            </div>
          )}

          {/* Current trick */}
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            dataTutorial="oh-trick-display"
            cardAriaLabelFor={(player, card) =>
              t('trickCardAriaLabel', {
                player: playerName(player.id, player.isHuman),
                card: cardAlt(card),
              })
            }
            winnerIdx={isTrickEnd ? state.leadPlayerIdx : undefined}
            winnerLabel={t('trickWinnerBadge')}
          />
        </div>

        {/* Right: info sidebar */}
        <div>
          {/* CPU players */}
          {isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                {tc('label.cpuOpponents', { count: state.players.filter((p) => !p.isHuman).length })}
              </summary>
              <div className="mt-1">
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                      {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                      {t('cumulativeScore', { score: p.cumulativeScore })} | {t('roundScore', { score: p.roundScore })}{' '}
                      | {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')}
                    </div>
                  ))}
              </div>
            </details>
          ) : (
            state.players
              .filter((p) => !p.isHuman)
              .map((p) => (
                <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                  <div className="text-ds-text-muted text-sm">
                    {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                    {t('cumulativeScore', { score: p.cumulativeScore })} | {t('roundScore', { score: p.roundScore })} |{' '}
                    {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')}
                  </div>
                </div>
              ))
          )}

          {/* Score table */}
          {isMobile ? (
            <details
              className="my-3 p-2 rounded bg-black/30 relative"
              data-tutorial="oh-score-table"
              open={isRoundEnd || isGameEnd || undefined}
            >
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('scores')}</summary>
              <div className="overflow-x-auto -mx-2 px-2">
                <table
                  className={scoreTableClass(isLargeDesktop, 'w-full text-sm text-ds-text-muted min-w-[360px]', 'mt-1')}
                >
                  <thead>
                    <tr>
                      <th scope="col" className="text-left">
                        {t('scoresPlayer')}
                      </th>
                      <th scope="col">{t('scoresBid')}</th>
                      <th scope="col" className="text-[10px] whitespace-nowrap">
                        {t('scoresTricks')}
                      </th>
                      <th scope="col" className="text-[10px] whitespace-nowrap">
                        {t('scoresRound')}
                      </th>
                      <th scope="col">{t('scoresTotal')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {state.players.map((p) => (
                      <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                        <td>{playerName(p.id, p.isHuman)}</td>
                        <td className="text-center">{p.bid >= 0 ? p.bid : '-'}</td>
                        <td className="text-center">{p.trickCount}</td>
                        <td className="text-center">{p.roundScore}</td>
                        <td className="text-center">{p.cumulativeScore}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              <ScrollFadeHint />
            </details>
          ) : (
            <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="oh-score-table">
              <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
              <div className="overflow-x-auto -mx-2 px-2">
                <table className={scoreTableClass(isLargeDesktop, 'w-full text-sm text-ds-text-muted min-w-[360px]')}>
                  <thead>
                    <tr>
                      <th scope="col" className="text-left">
                        {t('scoresPlayer')}
                      </th>
                      <th scope="col">{t('scoresBid')}</th>
                      <th scope="col" className="text-[10px] whitespace-nowrap">
                        {t('scoresTricks')}
                      </th>
                      <th scope="col" className="text-[10px] whitespace-nowrap">
                        {t('scoresRound')}
                      </th>
                      <th scope="col">{t('scoresTotal')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {state.players.map((p) => (
                      <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                        <td>{playerName(p.id, p.isHuman)}</td>
                        <td className="text-center">{p.bid >= 0 ? p.bid : '-'}</td>
                        <td className="text-center">{p.trickCount}</td>
                        <td className="text-center">{p.roundScore}</td>
                        <td className="text-center">{p.cumulativeScore}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
          <RoundScoreAnnouncement
            active={isRoundEnd || isGameEnd}
            entries={state.players.map((p) => ({
              name: `${playerName(p.id, p.isHuman)} ${t('scoreAnnouncementDetails', {
                bid: p.bid >= 0 ? p.bid : t('bidNone'),
                tricks: p.trickCount,
              })}`,
              roundScore: p.roundScore,
              cumulativeScore: p.cumulativeScore,
            }))}
          />
        </div>
      </div>

      {/* Message */}
      <div>{isMobile && <OhHellMessageBox state={state} />}</div>
    </>
  );
}

function FooterStatus({ game, t, frontendHint }: OhHellContext) {
  return (
    <>
      {/* ライブ領域は**常設**。hint がある間だけ現れる内側の要素に role/aria-live を
          付けると、領域と中身が同じコミットで DOM に入るので変化として扱われず、
          読み上げられないことがある (#5955, #6663)。 */}
      <div data-testid="ohhell-hint-live" role="status" aria-live="polite">
        {game.hint && (
          <div className="text-ds-warning text-sm mb-2">
            {game.hint.bid != null
              ? `${t('hintBid')}: ${game.hint.bid} (${t(`hintReason.${game.hint.reason}`)})`
              : `${t('hintPlay')}: [${game.hint.cardIndex}] (${t(`hintReason.${game.hint.reason}`)})`}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function primaryActions({ game, state, t, tc, busy }: OhHellContext) {
  const isBidPhase = state.phase === OhHellPhase.BID;
  const isPlayPhase = state.phase === OhHellPhase.PLAY;
  const isTrickEnd = state.phase === OhHellPhase.TRICK_END;
  const isRoundEnd = state.phase === OhHellPhase.ROUND_END;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;

  if (!isHumanBidTurn && !isHumanTurn && !isTrickEnd && !isRoundEnd) return null;

  return (
    <div className="flex gap-2 items-center" data-tutorial="oh-play-button">
      {(isHumanBidTurn || isHumanTurn) && (
        <button type="button" className={btnSuccess} onClick={game.handleHint} disabled={busy || game.hintLoading}>
          {tc('button.hint')}
        </button>
      )}
      {isHumanBidTurn && (
        <fieldset className="border-0 p-0 m-0 min-w-0 flex flex-wrap gap-1.5">
          <legend className="sr-only">{t('bidOptions', { max: state.handSize })}</legend>
          {Array.from({ length: state.handSize + 1 }, (_, i) => i).map((i) => {
            const isRestricted = state.restrictedBid === i;
            return (
              <button
                key={i}
                type="button"
                // Restricted bids use aria-disabled (not the HTML disabled
                // attribute) so they stay focusable and a screen reader can
                // read why they can't be chosen; the click is guarded instead.
                // Mirrors the Cribbage pegRestricted / Call Break pattern.
                // Neutralize btnPrimary's interactive feedback (press-scale,
                // hover shadow) when restricted so it doesn't feel clickable.
                // **ヒントは数字をテキストで言うだけで、どのボタンを
                // 押せばよいか視覚的に示していなかった。**制限ビッドとは
                // 別状態なので、強調は制限が無いときだけ付ける。
                className={`${btnPrimary}${
                  isRestricted ? ' opacity-50 cursor-not-allowed active:scale-100 hover:shadow-none' : ''
                }${!isRestricted && game.hint?.bid === i ? ' ring-2 ring-ds-warning' : ''}`}
                onClick={() => {
                  if (!isRestricted) game.handleBid(i);
                }}
                disabled={busy}
                aria-disabled={isRestricted || undefined}
                title={isRestricted ? t('restrictedBidTooltip') : undefined}
                aria-label={isRestricted ? t('restrictedBidAria', { n: i }) : t('bid', { n: i })}
                data-testid={isRestricted ? 'ohhell-restricted-bid' : undefined}
                data-hint-suggested={!isRestricted && game.hint?.bid === i ? 'true' : undefined}
              >
                {i}
              </button>
            );
          })}
        </fieldset>
      )}
      {isHumanTurn && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handlePlay}
          disabled={busy || game.selectedCardIndices.length !== 1}
        >
          {t('playButton')}
        </button>
      )}
      {isTrickEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextTrick} disabled={busy}>
          {t('nextTrick')}
        </button>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

function FooterHand({ game, state, t, view }: OhHellContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPlayPhase = state.phase === OhHellPhase.PLAY;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const { cardWidth, isMobile } = view;

  return (
    <>
      <span id="ohhell-illegal-card-description" className="sr-only">
        {t('illegalCardStatus')}
      </span>
      {/* Human cards */}
      {humanPlayer &&
        (isMobile ? (
          <MobileHandGrid
            cards={humanPlayer.cards}
            selectedIndices={game.selectedCardIndices}
            onToggle={game.toggleCard}
            cardWidth={cardWidth}
            stackRows
            dataTutorial="oh-player-hand"
            legalIndices={isHumanTurn ? state.validPlayIndices : undefined}
            validIndices={isHumanTurn ? state.validPlayIndices : undefined}
            cardStatusFor={(idx) =>
              isHumanTurn
                ? state.validPlayIndices.includes(idx)
                  ? t('legalCardStatus')
                  : t('illegalCardStatus')
                : undefined
            }
          />
        ) : (
          <div className="flex flex-wrap gap-1 mb-2" data-tutorial="oh-player-hand">
            {humanPlayer.cards.map((card, idx) => {
              const legal = !isHumanTurn || state.validPlayIndices.includes(idx);
              return (
                <button
                  type="button"
                  key={`${card.design}-${card.value}-${idx}`}
                  onClick={() => {
                    if (legal) game.toggleCard(idx);
                  }}
                  aria-label={`${cardAlt(card)}${isHumanTurn ? ` (${legal ? t('legalCardStatus') : t('illegalCardStatus')})` : ''}`}
                  aria-disabled={!legal || undefined}
                  aria-describedby={!legal ? 'ohhell-illegal-card-description' : undefined}
                  aria-pressed={game.selectedCardIndices.includes(idx)}
                  // プレイヒントの [N] もテキストだけだった。TwoTenJack と
                  // 同じく、該当する札を光らせる。
                  className={`transition-transform ${focusRingCard}${
                    isHumanTurn && legal ? ' rounded-lg ring-2 ring-ds-success' : ''
                  }${isHumanTurn && !legal ? ' opacity-50 cursor-not-allowed' : ''}${
                    isHumanTurn && game.hint?.cardIndex === idx ? ' rounded-lg ring-2 ring-ds-warning' : ''
                  }`}
                  data-hint-suggested={isHumanTurn && game.hint?.cardIndex === idx ? 'true' : undefined}
                  style={{
                    background: 'none',
                    padding: 0,
                    borderRadius: 8,
                    ...selectedCardStyle(game.selectedCardIndices.includes(idx)),
                    boxSizing: 'border-box',
                  }}
                >
                  <AnimatedCard card={card} width={cardWidth} />
                </button>
              );
            })}
          </div>
        ))}
    </>
  );
}

const ohHellDefinition: GameDefinition<OhHellResponse, OhHellGame, Parameters<typeof ohHellApi.exec>, OhHellView> = {
  key: 'ohhell',
  useGame: useOhHellGame,
  useView: useOhHellView,
  tutorial: OH_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 5 },
  cli: () => ({
    parseCommand: parseOhhellCommand,
    formatResponse: formatOhhellState,
    helpText: OHHELL_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === OhHellPhase.GAME_END || !!state.gameEndFlag,
  }),
  shell: ({ state }) => {
    const isBidPhase = state.phase === OhHellPhase.BID;
    const isPlayPhase = state.phase === OhHellPhase.PLAY;
    const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
    const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
    return {
      isHumanTurn: isHumanBidTurn || isHumanTurn,
    };
  },
  reset: ({ game }) => {
    void game.exec('reset', undefined, undefined, {
      cpuDifficulty: game.ohHellConfig.cpuDifficulty,
      maxHandSize: game.ohHellConfig.maxHandSize,
      scoringVariant: game.ohHellConfig.scoringVariant,
      roundDirection: game.ohHellConfig.roundDirection,
    });
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.ohHellConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'maxHandSize',
          label: t('settings.maxHandSize'),
          value: game.ohHellConfig.maxHandSize,
          options: MAX_HAND_SIZE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('maxHandSize', v),
        },
        {
          type: 'select',
          id: 'scoringVariant',
          label: t('settings.scoringVariant'),
          value: game.ohHellConfig.scoringVariant,
          options: SCORING_VARIANT_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.labelKey}`),
          })),
          onSelect: (v) => game.handleConfigChange('scoringVariant', v),
        },
        {
          type: 'select',
          id: 'roundDirection',
          label: t('settings.roundDirection'),
          value: game.ohHellConfig.roundDirection,
          options: ROUND_DIRECTION_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.labelKey}`),
          })),
          onSelect: (v) => game.handleConfigChange('roundDirection', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  message: (ctx) => !ctx.view.isMobile && <OhHellMessageBox state={ctx.state} />,
  cardNav: ({ game, state }) => {
    const isPlayPhase = state.phase === OhHellPhase.PLAY;
    const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
    return {
      cardCount: state.players.find((p) => p.isHuman)?.cards?.length ?? 0,
      onToggle: (idx) => {
        if (state.validPlayIndices.includes(idx)) game.toggleCard(idx);
      },
      onConfirm: game.handlePlay,
      onClear: game.clearSelection,
      enabled: isHumanTurn,
    };
  },
  shortcutsTestId: 'oh-hell-kbd-shortcuts',
  resetTutorialId: 'oh-reset-button',
  actionsClassName: 'flex items-center',
  HeaderExtra,
  Board,
  FooterHand,
  FooterStatus,
};

/** Renders the Oh Hell game page with bidding, trick play, and scoring. */
export const OhHellPage = createGamePage(ohHellDefinition);
