import { useEffect, useId, useMemo, useState } from 'react';
import type { dilotiApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { DILOTI_CPU_DIFFICULTY_OPTIONS, DILOTI_TARGET_OPTIONS, useDilotiGame } from '../hooks/useDilotiGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { DilotiResponse } from '../types/card';
import { DilotiPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { DILOTI_HELP, parseDilotiCommand } from '../utils/cli/commands/dilotiCommands';
import { formatDilotiState } from '../utils/cli/formatters/dilotiFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Diloti tutorial step definitions. */
const DILOTI_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="diloti-table"]',
    messageKey: 'tutorial.table',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="diloti-scores"]',
    messageKey: 'tutorial.scores',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="diloti-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="diloti-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="diloti-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

// **フェーズは文字列。** 共通の usePhaseNames は数値キーを前提にしている。
const DILOTI_PHASE_KEYS: Readonly<Record<string, string>> = {
  [DilotiPhase.PLAY]: 'play',
  [DilotiPhase.ROUND_END]: 'roundEnd',
  [DilotiPhase.GAME_END]: 'gameEnd',
};

type DilotiGame = ReturnType<typeof useDilotiGame>;
type DilotiView = ReturnType<typeof useDilotiView>;
type DilotiContext = GamePageContext<DilotiResponse, DilotiGame, DilotiView>;

function useDilotiView({ game, t }: { game: DilotiGame; t: DilotiContext['t'] }) {
  const { cardWidth, isMobile } = useCardDimensions();
  const [focusedTableIndices, setFocusedTableIndices] = useState<number[]>([]);
  const focusHelpId = useId();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const handleSelectHand = (idx: number) => {
    setFocusedTableIndices([]);
    game.selectHand(idx);
  };

  const phaseNames: Readonly<Record<string, string>> = useMemo(() => {
    const map: Record<string, string> = {};
    for (const [phase, key] of Object.entries(DILOTI_PHASE_KEYS)) {
      map[phase] = t(`phase.${key}`);
    }
    return map;
  }, [t]);

  return {
    cardWidth,
    isMobile,
    focusedTableIndices,
    setFocusedTableIndices,
    focusHelpId,
    handleSelectHand,
    phaseNames,
  };
}

function Board({ game, state, t, view, busy }: DilotiContext) {
  const { cardWidth, focusedTableIndices, setFocusedTableIndices, focusHelpId } = view;
  const lastCapturer = state.players.find((player) => player.id === state.lastCapturer);
  const isPlayPhase = state.phase === DilotiPhase.PLAY;
  const isRoundEnd = state.phase === DilotiPhase.ROUND_END;
  const isGameEnd = state.phase === DilotiPhase.GAME_END || state.gameEndFlag;
  const canPlay = isPlayPhase && state.isHumanTurn;
  const activeFocusedTableIndices = canPlay && game.selectedHandIdx !== null ? focusedTableIndices : [];

  // **打てる手はサーバが数えたものだけ。** 同ランク・合計一致・宣言が絡むので、
  // 画面側で組み直すと必ずずれる。
  const takes = game.selectedHandIdx !== null ? (state.takeOptions[game.selectedHandIdx] ?? []) : [];
  const declareCands = game.selectedHandIdx !== null ? (state.declareOptions[game.selectedHandIdx] ?? []) : [];
  const mayTrail = game.selectedHandIdx !== null && (state.canTrail[game.selectedHandIdx] ?? false);

  /** Renders a capture target list as "1, 2, decl 0". */
  const takeLabel = (tableIdxs: number[], declIdxs: number[]) =>
    [...tableIdxs.map(String), ...declIdxs.map((d) => t('declShort', { idx: d }))].join(', ');

  // **読み上げ名には札そのものを入れる (#7037)。** 見えている番号は CUI の
  // コマンドに打ち込む値でもあるので残すが、番号だけではどの札を取る手なのか
  // スクリーンリーダの利用者に伝わらない。宣言はまとめて取るので値で呼ぶ。
  const takeAriaLabel = (tableIdxs: number[], declIdxs: number[]) => {
    const parts = [
      // **索引はサーバが takeOptions で返したものをそのまま使う。** 場札にも
      // 宣言にも必ず対応があるので、「無かったとき」の分岐は置かない。
      ...tableIdxs.map((i) => cardAlt(state.table[i])),
      ...declIdxs.map((d) => t('declValue', { value: state.declarations[d].value })),
    ];
    return t('takeGroup', { cards: parts.join(', ') });
  };

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber, target: state.config.targetScore })}</span>
        <span data-testid="diloti-deck">{t('deck', { n: state.deckRemaining })}</span>
      </div>

      <div className={lgTwoColGrid}>
        <div data-tutorial="diloti-table">
          <div className="mb-1 text-ds-text-muted text-sm">{t('tableLabel')}</div>
          <div className="mb-2 p-2 rounded bg-black/30 flex flex-wrap gap-1" data-testid="diloti-table">
            {state.table.length === 0 ? (
              <span className="text-ds-text-muted text-sm">{t('tableEmpty')}</span>
            ) : (
              state.table.map((c, i) => (
                <div
                  key={`${c.design}-${c.value}-${i}`}
                  className={`flex flex-col items-center rounded ${activeFocusedTableIndices.includes(i) ? 'ring-2 ring-ds-warning' : ''}`}
                  data-testid={`diloti-table-card-${i}`}
                >
                  <CardImage card={c} width={cardWidth} />
                  <span className="text-xs text-ds-text-muted">{i}</span>
                </div>
              ))
            )}
          </div>

          {/* **宣言も番号付きで見せる。** 見えないと `d0` で取る対象を
              指せず、グループ宣言かどうかも分からない。 */}
          {state.declarations.length > 0 && (
            <div className="mb-2 p-2 rounded bg-black/30" data-testid="diloti-declarations">
              <div className="mb-1 text-ds-text-muted text-sm">{t('declarationsLabel')}</div>
              {state.declarations.map((d, i) => (
                <div
                  key={`${d.ownerIdx}-${d.value}-${i}`}
                  className="text-ds-text-primary text-sm py-0.5"
                  data-testid={`diloti-declaration-${i}`}
                >
                  {t('declarationLine', {
                    idx: i,
                    value: d.value,
                    kind: d.isGroup ? t('declGroup') : t('declPlain'),
                    owner: playerName(d.ownerIdx, d.ownerIdx === 0),
                  })}
                  <div className="flex flex-wrap gap-1 mt-0.5">
                    {d.groups.flat().map((c, j) => (
                      <CardImage key={`${c.design}-${c.value}-${j}`} card={c} width={cardWidth} />
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* **打てる手はサーバが数えたものだけ。** */}
          {canPlay && game.selectedHandIdx !== null && (
            <div className="mb-2 p-2 rounded bg-black/30" data-testid="diloti-move-options">
              <div className="text-ds-text-primary text-sm mb-1">
                {takes.length > 0 || declareCands.length > 0 ? t('chooseMove') : t('noCapture')}
              </div>
              <div className="flex flex-wrap gap-2">
                <span id={focusHelpId} className="sr-only">
                  {t('focusTableCards')}
                </span>
                {takes.map((o) => (
                  <button
                    key={`t-${o.tableIdxs.join('-')}-${o.declIdxs.join('-')}`}
                    type="button"
                    className={btnPrimary}
                    onClick={() => {
                      setFocusedTableIndices([]);
                      game.take(o.tableIdxs, o.declIdxs);
                    }}
                    disabled={busy}
                    data-testid={`diloti-take-${[...o.tableIdxs, ...o.declIdxs.map((d) => `d${d}`)].join('-')}`}
                    aria-label={takeAriaLabel(o.tableIdxs, o.declIdxs)}
                    aria-describedby={focusHelpId}
                    onFocus={() => setFocusedTableIndices(o.tableIdxs)}
                    onBlur={() => setFocusedTableIndices([])}
                    onMouseEnter={() => setFocusedTableIndices(o.tableIdxs)}
                    onMouseLeave={() => setFocusedTableIndices([])}
                  >
                    {t('takeGroup', { cards: takeLabel(o.tableIdxs, o.declIdxs) })}
                  </button>
                ))}
                {declareCands.map((c) => (
                  <button
                    key={`d-${c.value}-${c.tableIdxs.join('-')}`}
                    type="button"
                    className={btnSecondary}
                    onClick={() => {
                      setFocusedTableIndices([]);
                      game.declare(c.value, c.tableIdxs);
                    }}
                    disabled={busy}
                    data-testid={`diloti-declare-${c.value}-${c.tableIdxs.join('-')}`}
                    aria-label={t('declareGroup', {
                      value: c.value,
                      cards: c.tableIdxs.map((i) => cardAlt(state.table[i])).join(t('listSeparator')),
                    })}
                    aria-describedby={focusHelpId}
                    onFocus={() => setFocusedTableIndices(c.tableIdxs)}
                    onBlur={() => setFocusedTableIndices([])}
                    onMouseEnter={() => setFocusedTableIndices(c.tableIdxs)}
                    onMouseLeave={() => setFocusedTableIndices([])}
                  >
                    {t('declareGroup', { value: c.value, cards: c.tableIdxs.join(', ') })}
                  </button>
                ))}
                {mayTrail && (
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={game.trail}
                    disabled={busy}
                    data-testid="diloti-lay-off"
                  >
                    {t('layOffButton')}
                  </button>
                )}
              </div>
            </div>
          )}
        </div>

        <div data-tutorial="diloti-scores">
          <div className="mb-2 p-2 rounded bg-black/30" data-testid="diloti-scores">
            {state.players.map((p) => (
              <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                <span className="flex items-center gap-2">
                  <span>
                    {playerName(p.id, p.isHuman)}: {t('score', { n: p.score })} / {t('taken', { n: p.capturedCount })} /{' '}
                    {t('xeri', { n: p.xeri })}
                  </span>
                  {p.isDealer && (
                    <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('dealerBadge')}</span>
                  )}
                </span>
              </div>
            ))}
          </div>

          {state.lastResult && (isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="diloti-round-result">
              <div className="mb-1 text-ds-text-primary">{t('roundResult')}</div>
              {state.lastResult.lines
                .filter((l) => l.points[0] !== 0 || l.points[1] !== 0)
                .map((l) => (
                  <div key={l.key}>
                    {t(`score.${l.key}`)}: {l.points[0]} - {l.points[1]}
                  </div>
                ))}
              <div className="text-ds-text-primary">
                {t('roundTotal', { a: state.lastResult.totals[0], b: state.lastResult.totals[1] })}
              </div>
              {lastCapturer && (
                <div>{t('lastCapturer', { name: playerName(lastCapturer.id, lastCapturer.isHuman) })}</div>
              )}
              {state.players.map((player, idx) => (
                <div key={player.id}>
                  {t('roundScoreSummary', {
                    name: playerName(player.id, player.isHuman),
                    round: state.lastResult?.totals[idx] ?? 0,
                    total: player.score,
                  })}
                </div>
              ))}
              {!isGameEnd && <div className="text-ds-text-primary">{t('gameContinues')}</div>}
            </div>
          )}

          {isGameEnd && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-primary" data-testid="diloti-winner">
              {t('winner', { name: playerName(state.winnerIdx, state.winnerIdx === 0) })}
            </div>
          )}
        </div>
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ game, state, t, view }: DilotiContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;

  const isPlayPhase = state.phase === DilotiPhase.PLAY;
  const canPlay = isPlayPhase && state.isHumanTurn;
  const handValidIndices = canPlay ? humanPlayer.cards.map((_, i) => i) : undefined;
  const moveKinds = canPlay
    ? state.takeOptions.map((takesForCard, i) => {
        const canTake = takesForCard.length > 0;
        const canDeclare = (state.declareOptions[i] ?? []).length > 0;
        const canLayOff = state.canTrail[i] ?? false;
        return canTake || canDeclare || canLayOff ? { canTake, canDeclare, canLayOff } : null;
      })
    : undefined;

  return (
    <div data-testid="diloti-move-markers">
      <PlayerHandSection
        humanPlayer={humanPlayer}
        selectedCardIndices={game.selectedHandIdx === null ? [] : [game.selectedHandIdx]}
        toggleCard={view.handleSelectHand}
        cardWidth={view.cardWidth}
        isMobile={view.isMobile}
        dataTutorialPrefix="diloti"
        validIndices={handValidIndices}
        legalIndices={handValidIndices}
        restrictedTooltip={t('restrictedTooltip')}
        cardBadgeFor={(idx) => {
          const kind = moveKinds?.[idx];
          if (!kind) return null;
          const labels = [
            kind.canTake ? t('action.capture') : '',
            kind.canDeclare ? t('action.declare') : '',
            kind.canLayOff ? t('action.trail') : '',
          ].filter(Boolean);
          return {
            glyph: [kind.canTake ? '取' : '', kind.canDeclare ? '宣' : '', kind.canLayOff ? '置' : '']
              .filter(Boolean)
              .join('/'),
            title: labels.join(' / '),
          };
        }}
      />
    </div>
  );
}

function FooterStatus({ state, t, frontendHint }: DilotiContext) {
  return (
    <>
      {/* ライブ領域は常設 (#5955)。 */}
      <div data-testid="diloti-hint-live" role="status" aria-live="polite">
        {isRequestedHint(state) && state.hintHandIdx >= 0 && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {t(`hint.${state.hintReason}`)} ([{state.hintHandIdx}]
            {` → ${t(`action.${state.hintAction}`)}`})
            {state.hintAction === 'capture' && (
              <span>
                {state.hintTableIdxs.length > 0 && (
                  <>
                    {' '}
                    {t('hintTargets.captureTable', {
                      cards: state.hintTableIdxs
                        .map((idx) => `[${idx}] ${cardAlt(state.table[idx])}`)
                        .join(t('listSeparator')),
                    })}
                  </>
                )}
                {state.hintDeclIdxs.map((idx) => (
                  <span key={idx}>
                    {' '}
                    {t('hintTargets.captureDeclaration', { idx, value: state.declarations[idx].value })}
                  </span>
                ))}
              </span>
            )}
            {state.hintAction === 'declare' && (
              <span>
                {' '}
                {t('hintTargets.declare', {
                  value: state.hintDeclValue,
                  cards: state.hintTableIdxs
                    .map((idx) => `[${idx}] ${cardAlt(state.table[idx])}`)
                    .join(t('listSeparator')),
                })}
              </span>
            )}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ game, state, t, busy }: DilotiContext) {
  const isRoundEnd = state.phase === DilotiPhase.ROUND_END;
  return (
    <>
      {isRoundEnd && (
        <button
          type="button"
          className={btnSuccess}
          onClick={game.handleNextRound}
          disabled={busy}
          data-testid="diloti-next-round"
        >
          {t('nextRound')}
        </button>
      )}
    </>
  );
}

const dilotiDefinition: GameDefinition<DilotiResponse, DilotiGame, Parameters<typeof dilotiApi.exec>, DilotiView> = {
  key: 'diloti',
  useGame: useDilotiGame,
  useView: useDilotiView,
  tutorial: DILOTI_TUTORIAL_STEPS,
  skeleton: { kind: 'casino-table', sections: [4], footerStyle: 'hand', footerHandSize: 6 },
  cli: () => ({
    parseCommand: parseDilotiCommand,
    formatResponse: formatDilotiState,
    helpText: DILOTI_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === DilotiPhase.GAME_END || state.gameEndFlag;
    return {
      name: view.phaseNames[state.phase] ?? view.phaseNames[DilotiPhase.PLAY],
      ended: isGameEnd,
      won: isGameEnd && state.winnerIdx === 0,
    };
  },
  shell: ({ state }) => ({
    isHumanTurn: state.isHumanTurn,
  }),
  reset: ({ game }) => {
    game.reset();
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.dilotiConfig.cpuDifficulty,
          options: DILOTI_CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetScore',
          label: t('settings.targetScore'),
          value: game.dilotiConfig.targetScore,
          options: DILOTI_TARGET_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetScore', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'diloti-action-buttons',
  resetTutorialId: 'diloti-reset-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/**
 * Renders the Diloti page: the Greek fishing game where captures sum to the
 * played card's own rank, declarations build piles, and a one-card sweep of
 * the table is a xeri worth ten.
 */
export const DilotiPage = createGamePage(dilotiDefinition);
