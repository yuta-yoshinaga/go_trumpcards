import type { TFunction } from 'i18next';
import { useEffect, useMemo, useRef, useState } from 'react';
import type { sambaApi } from '../api/gameApi';
import { CardFan } from '../components/CardFan';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { MobileHandGrid } from '../components/MobileHandGrid';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { useBreakpoint, useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useSambaGame } from '../hooks/useSambaGame';
import { badgeInfoColors } from '../styles/badgeStyles';
import { btnOutline, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { Card, SambaResponse } from '../types/card';
import { SambaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { canastaFamilySelectionPoints } from '../utils/canastaFamilyScore';
import { cardAlt } from '../utils/cardAlt';
import { parseSambaCommand, SAMBA_HELP } from '../utils/cli/commands/sambaCommands';
import { formatSambaState } from '../utils/cli/formatters/sambaFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/**
 * Cards required to complete a canasta (7-card set) or samba (7-card sequence).
 * Matches the Go domain rule (`SambaMeld.IsCanasta`/`IsSamba`: `len(Cards) >= 7`).
 */
const SAMBA_CANASTA_SIZE = 7;

const SAMBA_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SambaPhase.DRAW]: 'draw',
  [SambaPhase.MELD]: 'meld',
  [SambaPhase.DISCARD]: 'discard',
  [SambaPhase.ROUND_END]: 'roundEnd',
  [SambaPhase.GAME_END]: 'gameEnd',
};

/** Samba tutorial step definitions. */
const SA_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sa-draw-area"]', messageKey: 'tutorial.drawArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="sa-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="sa-meld-area"]', messageKey: 'tutorial.meldArea', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="sa-actions"]', messageKey: 'tutorial.actionButtons', placement: 'top', advanceOn: 'next' },
];

function useGame() {
  const game = useSambaGame();
  return {
    ...game,
    exec: game.gameExec,
  };
}

type SambaGame = ReturnType<typeof useGame>;
type SambaView = ReturnType<typeof useSambaView>;
type SambaContext = GamePageContext<SambaResponse, SambaGame, SambaView>;

function useSambaView({ game, t }: { game: SambaGame; t: TFunction }) {
  const isMobile = useIsMobile();
  const { selectedCardIndices } = game;
  const state = game.state;

  const { cardWidth } = useCardDimensions();
  const isLargeDesktop = useBreakpoint() === 'largeDesktop';
  const phaseNames = usePhaseNames('samba', SAMBA_PHASE_KEYS);

  const humanPlayer = state?.players.find((p) => p.isHuman);
  const humanCardCount = humanPlayer?.cards?.length ?? 0;
  const lastHandCard = humanPlayer?.cards[0];
  const humanHasLastRed3 =
    humanCardCount === 1 &&
    lastHandCard?.value === 3 &&
    (lastHandCard.design === 'HEART' || lastHandCard.design === 'DIAMOND');
  const isDrawPhase = state?.phase === SambaPhase.DRAW;
  const isMeldPhase = state?.phase === SambaPhase.MELD;
  const isDiscardPhase = state?.phase === SambaPhase.DISCARD;
  const completedMelds = state && humanPlayer ? state.completedMelds[humanPlayer.team] : 0;
  const goOutRequiredMelds = state?.config.goOutRequiredMelds ?? 0;
  const canGoOut = humanCardCount <= 1 && !humanHasLastRed3 && completedMelds >= goOutRequiredMelds;
  const goOutUnavailableReason =
    humanCardCount > 1
      ? t('goOutHandTooLarge')
      : humanHasLastRed3
        ? t('goOutRedThree')
        : completedMelds < goOutRequiredMelds
          ? t('goOutUnavailable', { required: goOutRequiredMelds })
          : '';

  const drawDiscardReason = useMemo(() => {
    if (!isDrawPhase) return '';
    const n = selectedCardIndices.length;
    if (n > 2) return t('drawDiscardReason.tooMany');
    if (n === 2) return '';
    // Frozen takes priority while the player is still picking — the wildcard restriction
    // is the load-bearing rule players forget; surface it whether they've picked 0 or 1 cards.
    if (state?.isFrozen) return t('drawDiscardReason.frozen');
    if (n === 1) return t('drawDiscardReason.selectOneMore');
    return t('drawDiscardReason.selectTwo');
  }, [isDrawPhase, selectedCardIndices.length, state?.isFrozen, t]);

  // Meld phase: surface the initial-meld minimum (by team score band) and the
  // selected cards' running point total so the player can tell if they qualify.
  const meldPointInfo = useMemo(() => {
    if (!isMeldPhase || !humanPlayer) return null;
    const selectedCards = selectedCardIndices.map((i) => humanPlayer.cards[i]).filter((c): c is Card => Boolean(c));
    const selectedPoints = canastaFamilySelectionPoints(selectedCards);
    const needInitial = !humanPlayer.hasInitMeld;
    const minMeld = state.minMeld;
    const shortfall = Math.max(0, minMeld - selectedPoints);
    return { selectedPoints, needInitial, minMeld, shortfall, below: needInitial && shortfall > 0 };
  }, [isMeldPhase, humanPlayer, selectedCardIndices, state?.minMeld]);

  const isHumanTurn =
    (isDrawPhase || isMeldPhase || isDiscardPhase) && state?.players[state.currentPlayerIdx]?.isHuman === true;

  const [meldProgressMsg, setMeldProgressMsg] = useState('');
  const [teamScoreAnnouncement, setTeamScoreAnnouncement] = useState('');
  const previousTeamScores = useRef<[number, number] | null>(null);
  useEffect(() => {
    if (!state) return;
    const scores: [number, number] = [state.teamScores[0] ?? 0, state.teamScores[1] ?? 0];
    const previousScores = previousTeamScores.current;
    previousTeamScores.current = scores;
    if (!previousScores) return;

    const updates = scores.flatMap((score, team) =>
      score !== previousScores[team] ? [t('teamScoreAnnouncement', { team: t('teamLabel', { n: team }), score })] : [],
    );
    if (updates.length > 0) {
      setTeamScoreAnnouncement(updates.join(t('listSeparator')));
    }
  }, [state, t]);

  const previousHumanTeamMelds = useRef<Map<string, number> | null>(null);
  useEffect(() => {
    if (!state || !humanPlayer) return;
    const teamMelds = state.players
      .filter((player) => player.team === humanPlayer.team)
      .flatMap((player) => player.melds.map((meld, index) => ({ key: `${player.id}-${index}`, meld })));
    const currentLengths = new Map(teamMelds.map(({ key, meld }) => [key, meld.cards.length]));
    const previousLengths = previousHumanTeamMelds.current;
    previousHumanTeamMelds.current = currentLengths;
    if (!previousLengths) return;

    const updates = teamMelds
      .filter(({ key, meld }) => meld.cards.length > (previousLengths.get(key) ?? 0))
      .map(({ meld }) => {
        const remaining = SAMBA_CANASTA_SIZE - meld.cards.length;
        return remaining <= 0
          ? t(meld.kind === 1 ? 'meldProgress.sambaComplete' : 'meldProgress.canastaComplete')
          : t(meld.kind === 1 ? 'meldProgress.toSamba' : 'meldProgress.toCanasta', { n: remaining });
      });
    if (updates.length > 0) {
      setMeldProgressMsg(t('meldProgressAnnouncement', { updates: updates.join(t('listSeparator')) }));
    }
  }, [state, humanPlayer, t]);

  // Announce freeze/thaw transitions: freezing changes the draw rules (you now
  // need two matching cards to take the discard pile), so it must reach SR users
  // beyond the colour/badge cue.
  const [frozenMsg, setFrozenMsg] = useState('');
  const prevFrozenRef = useRef<boolean | null>(null);
  useEffect(() => {
    const frozen = state?.isFrozen;
    if (frozen == null) return;
    const prev = prevFrozenRef.current;
    prevFrozenRef.current = frozen;
    if (prev != null && prev !== frozen) {
      setFrozenMsg(frozen ? t('frozenAnnounceOn') : t('frozenAnnounceOff'));
      const id = setTimeout(() => setFrozenMsg(''), 3000);
      return () => clearTimeout(id);
    }
  }, [state?.isFrozen, t]);

  return {
    isMobile,
    cardWidth,
    isLargeDesktop,
    phaseNames,
    humanPlayer,
    canGoOut,
    completedMelds,
    goOutRequiredMelds,
    goOutUnavailableReason,
    drawDiscardReason,
    meldPointInfo,
    isHumanTurn,
    frozenMsg,
    meldProgressMsg,
    teamScoreAnnouncement,
  };
}

/** Human team's meld boxes (also shown beside the hand on large desktop). */
function humanTeamMeldsOf({ state, t, view }: SambaContext) {
  const { isLargeDesktop, cardWidth } = view;
  return state.players
    .filter((p) => p.team === state.players[0]?.team)
    .map((p, pi) => {
      if (p.melds.length === 0 && p.red3s.length === 0) return null;
      return (
        <div
          key={p.id}
          className={`${isLargeDesktop ? 'my-1 shrink-0 max-w-full p-1' : 'my-2 p-2'} rounded bg-ds-surface`}
          data-tutorial={p.isHuman ? 'sa-meld-area' : undefined}
        >
          <div className="text-ds-text-muted text-sm mb-1">
            {playerName(p.id, p.isHuman)} ({t('teamLabel', { n: p.team })}) - {t('melds')}
            {p.hasCanasta && <span className="ml-2 text-ds-warning">★</span>}
            {p.hasSamba && <span className="ml-1 text-ds-accent">▲</span>}
          </div>
          {p.melds.map((m, mi) => {
            // Progress toward the 7-card canasta (set) / samba (sequence)
            // milestone, derived purely from `cards.length` and `kind`.
            const remaining = SAMBA_CANASTA_SIZE - m.cards.length;
            const toSamba = m.kind === 1;
            const complete = remaining <= 0;
            return (
              <div key={mi} className={`flex ${isLargeDesktop ? 'flex-nowrap' : 'flex-wrap'} gap-1 mb-1`}>
                <span className="text-xs text-ds-text-muted self-center mr-1">
                  {m.kind === 1
                    ? m.isSamba
                      ? t('samba')
                      : t('sequence')
                    : m.isCanasta
                      ? m.isNatural
                        ? t('naturalCanasta')
                        : t('mixedCanasta')
                      : `(${m.cards.length})`}
                </span>
                <span
                  data-testid={`sa-meld-progress-${pi}-${mi}`}
                  className={`text-xs self-center mr-1 ${
                    complete ? 'text-ds-success font-bold motion-safe:animate-pulse' : 'text-ds-info'
                  }`}
                >
                  {complete
                    ? t(toSamba ? 'meldProgress.sambaComplete' : 'meldProgress.canastaComplete')
                    : t(toSamba ? 'meldProgress.toSamba' : 'meldProgress.toCanasta', { n: remaining })}
                </span>
                {m.cards.map((card, ci) => (
                  <AnimatedCard
                    key={`meld-${pi}-${mi}-${ci}`}
                    card={card}
                    width={cardWidth * 0.6}
                    wrapperClassName={isLargeDesktop ? `inline-block ${ci > 0 ? '-ml-12' : ''}` : undefined}
                  />
                ))}
              </div>
            );
          })}
          {p.red3s.length > 0 && (
            <div className={`flex ${isLargeDesktop ? 'flex-nowrap' : 'flex-wrap'} gap-1 mt-1`}>
              <span className="text-xs text-ds-error-text self-center mr-1">{t('red3s')}</span>
              {p.red3s.map((card, ri) => (
                <AnimatedCard
                  key={`red3-${pi}-${ri}`}
                  card={card}
                  width={cardWidth * 0.6}
                  wrapperClassName={isLargeDesktop ? `inline-block ${ri > 0 ? '-ml-12' : ''}` : undefined}
                />
              ))}
            </div>
          )}
        </div>
      );
    });
}

function handCardRenderer({ game, view }: SambaContext) {
  const { selectedCardIndices, meldGroups, toggleCard } = game;
  const { cardWidth, isHumanTurn, meldPointInfo } = view;
  const renderHandCard = (card: Card, idx: number) => {
    const isSelected = selectedCardIndices.includes(idx);
    const isGrouped = meldGroups.some((group) => group.includes(idx));
    return (
      <button
        type="button"
        key={`${card.design}-${card.value}-${idx}`}
        onClick={() => {
          if (!isGrouped) toggleCard(idx);
        }}
        aria-label={cardAlt(card)}
        aria-pressed={isSelected || isGrouped}
        aria-disabled={isGrouped || undefined}
        aria-describedby={
          [
            isGrouped ? 'sa-meld-card-assigned' : undefined,
            isHumanTurn && meldPointInfo?.needInitial ? 'sa-meld-points' : undefined,
          ]
            .filter(Boolean)
            .join(' ') || undefined
        }
        className={`transition-transform ${focusRingCard}`}
        style={{
          background: 'none',
          padding: 0,
          borderRadius: 8,
          ...selectedCardStyle(isSelected || isGrouped),
          boxSizing: 'border-box',
        }}
      >
        <AnimatedCard card={card} width={cardWidth} />
      </button>
    );
  };
  return renderHandCard;
}

function Board(ctx: SambaContext) {
  const { state, t, tc, frontendHint, view } = ctx;
  const { isMobile, isLargeDesktop, cardWidth, isHumanTurn, canGoOut, completedMelds, goOutRequiredMelds } = view;
  const { goOutUnavailableReason, frozenMsg, meldProgressMsg, teamScoreAnnouncement } = view;
  const isDiscardPhase = state.phase === SambaPhase.DISCARD;
  const isRoundEnd = state.phase === SambaPhase.ROUND_END;
  const isGameEnd = state.phase === SambaPhase.GAME_END || !!state.gameEndFlag;
  const humanTeamMelds = humanTeamMeldsOf(ctx);
  const messageBox = (
    <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
  );

  return (
    <>
      {isDiscardPhase && isHumanTurn && (
        <>
          <div
            className={`w-full text-xs ${canGoOut ? 'text-ds-success' : 'text-ds-warning'}`}
            data-testid="sa-go-out-progress"
          >
            {t('goOutProgress', { completed: completedMelds, required: goOutRequiredMelds })}
          </div>
          <div id="sa-go-out-guidance" className="w-full text-xs text-ds-warning" data-testid="sa-go-out-guidance">
            {goOutUnavailableReason}
          </div>
        </>
      )}
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span>
          {t('drawPile', { count: state.drawPileCount })} / {t('discardPile', { count: state.discardPileCount })}
        </span>
        {state.isFrozen && <span className="ml-2 whitespace-nowrap text-ds-info font-bold">[{t('frozen')}]</span>}
        {/* Self-contained live region announcing freeze/thaw transitions. */}
        <span className="sr-only" role="status" aria-live="polite" data-testid="sa-frozen-announce">
          {frozenMsg}
        </span>
        <span className="sr-only" role="status" aria-live="polite" data-testid="sa-meld-progress-announce">
          {meldProgressMsg}
        </span>
        <span className="sr-only" role="status" aria-live="polite" data-testid="sa-team-score-announce">
          {teamScoreAnnouncement}
        </span>
      </div>
      <div className="text-ds-text-muted text-center mb-2 text-sm" data-testid="sa-team-scores">
        {t('teamScores', { a: state.teamScores[0] ?? 0, b: state.teamScores[1] ?? 0 })}
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: game play area */}
        <div>
          {/* Discard pile top */}
          {state.discardTop && (
            <div
              className={`my-3 p-3 rounded flex items-center gap-3 relative ${
                state.isFrozen ? `${badgeInfoColors} ring-2 ring-ds-info` : 'bg-black/40'
              }`}
              data-tutorial="sa-draw-area"
              data-testid="sa-discard-pile"
            >
              <AnimatedCard card={state.discardTop} width={cardWidth} />
              <div className="text-ds-text-muted text-sm">{t('discardTop')}</div>
              {state.isFrozen && (
                <span
                  className="absolute top-1 right-2 shrink-0 whitespace-nowrap text-ds-info text-xs font-bold"
                  data-testid="sa-frozen-badge"
                  role="img"
                  aria-label={t('frozenIndicator')}
                >
                  {t('frozenIndicator')}
                </span>
              )}
            </div>
          )}

          {/* Human-team melds stay visible; opponents' melds are available on demand. */}
          {!isLargeDesktop && humanTeamMelds}
          {state.players.some(
            (p) => p.team !== state.players[0]?.team && (p.melds.length > 0 || p.red3s.length > 0),
          ) && (
            <details
              data-testid="sa-opponent-meld-details"
              open={!isMobile || undefined}
              className="my-2 p-2 rounded bg-ds-surface"
            >
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                {t('opponentMelds', {
                  count: state.players
                    .filter((p) => p.team !== state.players[0]?.team)
                    .reduce((count, p) => count + p.melds.length, 0),
                })}
              </summary>
              {state.players.map((p, pi) => {
                if (p.team === state.players[0]?.team || (p.melds.length === 0 && p.red3s.length === 0)) return null;
                return (
                  <div key={p.id} className="my-2 p-2 rounded bg-ds-surface">
                    <div className="text-ds-text-muted text-sm mb-1">
                      {playerName(p.id, p.isHuman)} ({t('teamLabel', { n: p.team })}) - {t('melds')}
                      {p.hasCanasta && <span className="ml-2 text-ds-warning">★</span>}
                      {p.hasSamba && <span className="ml-1 text-ds-accent">▲</span>}
                    </div>
                    {p.melds.map((m, mi) => {
                      const remaining = SAMBA_CANASTA_SIZE - m.cards.length;
                      const toSamba = m.kind === 1;
                      const complete = remaining <= 0;
                      return (
                        <div key={mi} className="flex flex-wrap gap-1 mb-1">
                          <span className="text-xs text-ds-text-muted self-center mr-1">
                            {m.kind === 1
                              ? m.isSamba
                                ? t('samba')
                                : t('sequence')
                              : m.isCanasta
                                ? m.isNatural
                                  ? t('naturalCanasta')
                                  : t('mixedCanasta')
                                : `(${m.cards.length})`}
                          </span>
                          <span
                            className={`text-xs self-center mr-1 ${complete ? 'text-ds-success font-bold motion-safe:animate-pulse' : 'text-ds-info'}`}
                          >
                            {complete
                              ? t(toSamba ? 'meldProgress.sambaComplete' : 'meldProgress.canastaComplete')
                              : t(toSamba ? 'meldProgress.toSamba' : 'meldProgress.toCanasta', { n: remaining })}
                          </span>
                          {m.cards.map((card, ci) => (
                            <AnimatedCard key={`opponent-meld-${pi}-${mi}-${ci}`} card={card} width={cardWidth * 0.6} />
                          ))}
                        </div>
                      );
                    })}
                    {p.red3s.length > 0 && (
                      <div className="flex flex-wrap gap-1 mt-1">
                        <span className="text-xs text-ds-error-text self-center mr-1">{t('red3s')}</span>
                        {p.red3s.map((card, ri) => (
                          <AnimatedCard key={`opponent-red3-${pi}-${ri}`} card={card} width={cardWidth * 0.6} />
                        ))}
                      </div>
                    )}
                  </div>
                );
              })}
            </details>
          )}
        </div>

        {/* Right: info sidebar */}
        <div>
          {/* Score table */}
          <details
            data-testid="sa-score-details"
            open={!isMobile || undefined}
            className="my-3 p-2 rounded bg-black/30"
          >
            <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('score.title')}</summary>
            <table className="w-full text-sm text-ds-text-muted">
              <thead>
                <tr>
                  <th scope="col" className="text-left">
                    {tc('label.player')}
                  </th>
                  <th scope="col">{t('score.team')}</th>
                  <th scope="col">{t('score.round')}</th>
                  <th scope="col">{t('score.cumulative')}</th>
                </tr>
              </thead>
              <tbody>
                {state.players.map((p) => (
                  <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                    <td>{playerName(p.id, p.isHuman)}</td>
                    <td className="text-center">{p.team}</td>
                    <td className="text-center">{p.roundScore}</td>
                    <td className="text-center">{p.cumulativeScore}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </details>

          {/* CPU hands (shown at round/game end) */}
          {(isRoundEnd || isGameEnd) && (
            <CpuAccordion playerCount={state.players.filter((p) => !p.isHuman).length}>
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                    <div className="text-ds-text-muted text-sm">
                      {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })}
                    </div>
                    {p.cards.length > 0 && (
                      <div className="flex flex-wrap gap-1 mt-1">
                        {p.cards.map((card, idx) => (
                          <AnimatedCard
                            key={`cpu-${card.design}-${card.value}-${idx}`}
                            card={card}
                            width={cardWidth * 0.7}
                          />
                        ))}
                      </div>
                    )}
                  </div>
                ))}
            </CpuAccordion>
          )}
        </div>
      </div>

      {isMobile && messageBox}

      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterHand(ctx: SambaContext) {
  const { game, t, view } = ctx;
  const { selectedCardIndices, meldGroups, toggleCard } = game;
  const { isMobile, isLargeDesktop, cardWidth, humanPlayer, isHumanTurn, meldPointInfo } = view;
  const visibleHumanTeamMelds = humanTeamMeldsOf(ctx).filter((meld) => meld !== null);
  const renderHandCard = handCardRenderer(ctx);

  return (
    <>
      <div
        className={isLargeDesktop ? 'flex items-start gap-4' : undefined}
        data-testid={isLargeDesktop ? 'sa-footer-hand-row' : undefined}
      >
        {isLargeDesktop && visibleHumanTeamMelds.length > 0 && (
          <div className="flex flex-wrap shrink-0 items-start gap-2 max-w-[30%]" data-testid="sa-footer-melds">
            {visibleHumanTeamMelds}
          </div>
        )}
        {humanPlayer &&
          (isMobile ? (
            <MobileHandGrid
              cards={humanPlayer.cards}
              selectedIndices={[...selectedCardIndices, ...meldGroups.flat()]}
              onToggle={(idx) => {
                if (!meldGroups.some((group) => group.includes(idx))) toggleCard(idx);
              }}
              cardWidth={cardWidth}
              dataTutorial="sa-player-hand"
              validIndices={humanPlayer.cards.flatMap((_, idx) =>
                meldGroups.some((group) => group.includes(idx)) ? [] : [idx],
              )}
              stackRows
            />
          ) : isLargeDesktop ? (
            <CardFan itemWidth={cardWidth + 6} dataTutorial="sa-player-hand" data-testid="sa-player-hand">
              {humanPlayer.cards.map(renderHandCard)}
            </CardFan>
          ) : (
            <div className="flex flex-wrap gap-1 mb-2" data-tutorial="sa-player-hand">
              {humanPlayer.cards.map(renderHandCard)}
            </div>
          ))}
      </div>

      <div id="sa-meld-points" role="status" aria-live="polite" data-testid="sa-meld-points">
        {isHumanTurn && meldPointInfo && (
          <div className={`w-full text-xs ${meldPointInfo.below ? 'text-ds-warning' : 'text-ds-text-muted'}`}>
            {meldPointInfo.needInitial
              ? t('meldPoints.initial', {
                  min: meldPointInfo.minMeld,
                  points: meldPointInfo.selectedPoints,
                  shortfall: meldPointInfo.below
                    ? t('meldPoints.shortfall', { points: meldPointInfo.shortfall })
                    : t('meldPoints.met'),
                })
              : t('meldPoints.selected', { points: meldPointInfo.selectedPoints })}
          </div>
        )}
      </div>
    </>
  );
}

function primaryActions({ game, state, t, busy: loading, view }: SambaContext) {
  const isDrawPhase = state.phase === SambaPhase.DRAW;
  const isMeldPhase = state.phase === SambaPhase.MELD;
  const isDiscardPhase = state.phase === SambaPhase.DISCARD;
  const isRoundEnd = state.phase === SambaPhase.ROUND_END;
  const { selectedCardIndices, meldGroups } = game;
  const { handleDrawStock, handleDrawDiscard, handleRemoveMeldGroup, handleAddMeldGroup, handleMeldSelected } = game;
  const { handleSkipMeld, handleDiscard, handleGoOut, handleNextRound } = game;
  const { isHumanTurn, drawDiscardReason, meldPointInfo, canGoOut, goOutUnavailableReason } = view;

  return isHumanTurn || isRoundEnd ? (
    <div className="flex gap-2 items-center flex-wrap" data-tutorial="sa-actions">
      {isDrawPhase && isHumanTurn && (
        <div className="flex gap-2 flex-col">
          <div className="flex gap-2">
            <button type="button" className={btnPrimary} onClick={handleDrawStock} disabled={loading}>
              {t('drawStockButton')}
            </button>
            <button
              type="button"
              className={btnPrimary}
              onClick={handleDrawDiscard}
              disabled={loading || selectedCardIndices.length !== 2}
              title={drawDiscardReason || undefined}
              aria-describedby={drawDiscardReason ? 'sa-draw-discard-reason' : undefined}
            >
              {t('drawDiscardButton')}
            </button>
          </div>
          {drawDiscardReason && (
            <div
              id="sa-draw-discard-reason"
              data-testid="sa-draw-discard-reason"
              className="text-xs text-ds-text-muted"
            >
              {drawDiscardReason}
            </div>
          )}
        </div>
      )}
      {isMeldPhase && isHumanTurn && (
        <>
          {meldGroups.length > 0 && (
            <section aria-label={t('meldGroups.label')} className="w-full flex flex-wrap gap-2">
              <span id="sa-meld-card-assigned" className="sr-only">
                {t('meldGroups.cardAssigned')}
              </span>
              {meldGroups.map((group, groupIndex) => (
                <div
                  key={`${groupIndex}-${group.join('-')}`}
                  className="flex items-center gap-2 rounded-md bg-ds-surface px-2 py-1"
                >
                  <span>
                    {t('meldGroups.group', { n: groupIndex + 1 })} ({t('meldGroups.cardCount', { count: group.length })}
                    )
                  </span>
                  <button
                    type="button"
                    className={`${btnOutline} aria-disabled:opacity-50 aria-disabled:cursor-not-allowed`}
                    onClick={() => {
                      if (!loading) handleRemoveMeldGroup(groupIndex);
                    }}
                    aria-disabled={loading || undefined}
                  >
                    {t('meldGroups.remove')}
                  </button>
                </div>
              ))}
            </section>
          )}
          <button
            type="button"
            className={`${btnOutline} aria-disabled:opacity-50 aria-disabled:cursor-not-allowed`}
            onClick={() => {
              if (loading || selectedCardIndices.length < 3) return;
              handleAddMeldGroup();
            }}
            aria-disabled={loading || selectedCardIndices.length < 3 || undefined}
          >
            {t('meldGroups.add')}
          </button>
          <button
            type="button"
            className={`${btnPrimary} aria-disabled:opacity-50 aria-disabled:cursor-not-allowed`}
            onClick={() => {
              if (
                loading ||
                (meldGroups.length === 0 ? selectedCardIndices.length < 3 : selectedCardIndices.length > 0)
              )
                return;
              handleMeldSelected();
            }}
            aria-disabled={
              loading ||
              (meldGroups.length === 0 ? selectedCardIndices.length < 3 : selectedCardIndices.length > 0) ||
              undefined
            }
            aria-describedby={isHumanTurn && meldPointInfo?.needInitial ? 'sa-meld-points' : undefined}
          >
            {t('meldButton')}
          </button>
          <button type="button" className={btnOutline} onClick={handleSkipMeld} disabled={loading}>
            {t('skipMeldButton')}
          </button>
        </>
      )}
      {isDiscardPhase && isHumanTurn && (
        <>
          <button
            type="button"
            className={btnPrimary}
            onClick={handleDiscard}
            disabled={loading || selectedCardIndices.length !== 1}
          >
            {t('discardButton')}
          </button>
          <button
            type="button"
            className={`${btnSuccess} aria-disabled:cursor-not-allowed aria-disabled:opacity-50`}
            onClick={() => {
              if (!loading && canGoOut) handleGoOut();
            }}
            aria-disabled={loading || !canGoOut}
            aria-describedby={goOutUnavailableReason ? 'sa-go-out-guidance' : undefined}
            title={goOutUnavailableReason || undefined}
          >
            {t('goOutButton')}
          </button>
        </>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
          {t('nextRound')}
        </button>
      )}
    </div>
  ) : null;
}

const sambaDefinition: GameDefinition<SambaResponse, SambaGame, Parameters<typeof sambaApi.exec>, SambaView> = {
  key: 'samba',
  useGame,
  useView: useSambaView,
  tutorial: SA_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', opponents: 3, centerCard: true, trickArea: true, footerHandSize: 13 },
  cli: ({ frontendHint }) => ({
    parseCommand: parseSambaCommand,
    formatResponse: formatSambaState,
    helpText: SAMBA_HELP,
    localCommand: hintLocalCommand(frontendHint.hint),
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === SambaPhase.GAME_END || !!state.gameEndFlag,
  }),
  shell: ({ view }) => ({ isHumanTurn: view.isHumanTurn }),
  reset: ({ game }) => {
    void game.exec('reset', undefined, {
      cpuDifficulty: game.sambaConfig.cpuDifficulty,
      pointLimit: game.sambaConfig.pointLimit,
    });
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.sambaConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'pointLimit',
          label: t('settings.pointLimit'),
          value: game.sambaConfig.pointLimit,
          options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('pointLimit', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  actionsClassName: '',
  message: ({ state, view }) =>
    view.isMobile ? undefined : (
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    ),
  cardNav: ({ game, state, view }) => ({
    cardCount: view.humanPlayer?.cards?.length ?? 0,
    onToggle: game.toggleCard,
    onConfirm: () => {
      if (state.phase === SambaPhase.DISCARD) game.handleDiscard();
      else if (state.phase === SambaPhase.MELD) game.handleMeldSelected();
    },
    onClear: game.clearSelection,
    enabled: !!view.isHumanTurn,
  }),
  shortcutsTestId: 'samba-kbd-shortcuts',
  Board,
  FooterHand,
};

/** Samba game page. */
export const SambaPage = createGamePage(sambaDefinition);
