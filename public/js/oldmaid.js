$(document).ready(function () {

    // --- ボタンイベント ---

    $('#btn-reset').on('click', function () {
        oldmaidRequest({ command: 'reset' });
    });

    $('#btn-draw').on('click', function () {
        oldmaidRequest({ command: 'draw' });
    });

    // --- API通信 ---

    function oldmaidRequest(payload) {
        $.ajax({
            url: './oldmaid/exec',
            type: 'POST',
            contentType: 'application/json',
            data: JSON.stringify(payload),
        }).done(function (response) {
            console.log(response);
            updateUi(response);
        }).fail(function () {
            console.log('oldmaid request failed');
        });
    }

    // --- UI更新 ---

    function updateUi(response) {
        if (!response || !response.players) return;

        response.players.forEach(function (player) {
            updatePlayerArea(player, response);
        });

        updateStatus(response);
        updateResult(response);
        updateButtons(response);
    }

    function updatePlayerArea(player, response) {
        var id = player.id;
        var $area = $('#player-area-' + id);
        var $label = $('#label-' + id);
        var $count = $('#count-' + id);
        var $cards = $('#cards-' + id);

        // Finished / draw-target styling
        $area.removeClass('draw-target finished-area');
        if (player.isFinished) {
            $area.addClass('finished-area');
        } else if (!response.gameEndFlag && response.nextDrawTargetIdx === id) {
            $area.addClass('draw-target');
        }

        // Badges in label
        $label.find('.finished-badge, .draw-target-badge').remove();
        if (player.isFinished) {
            $label.append('<span class="finished-badge">上がり</span>');
        } else if (!response.gameEndFlag && response.nextDrawTargetIdx === id && !player.isHuman) {
            $label.append('<span class="draw-target-badge">← 引く相手</span>');
        }

        // Card count label
        if (player.isFinished) {
            $count.text('');
        } else {
            $count.text(player.cardCount + '枚');
        }

        // Render cards
        $cards.empty();
        if (player.isFinished) {
            return;
        }

        if (player.isHuman) {
            // Show actual cards
            if (player.cards) {
                player.cards.forEach(function (card) {
                    var $img = $('<img>').attr('src', getImagePath(card)).attr('alt', cardLabel(card));
                    var $wrap = $('<div class="card-wrap"></div>').append($img);
                    $cards.append($wrap);
                });
            }
        } else {
            // Show card backs (max 10 visible)
            var showCount = Math.min(player.cardCount, 10);
            for (var i = 0; i < showCount; i++) {
                var $img = $('<img>').attr('src', './images/z01.png').attr('alt', 'card');
                var $wrap = $('<div class="card-wrap"></div>').append($img);
                $cards.append($wrap);
            }
            if (player.cardCount > 10) {
                $cards.append('<span class="card-count-extra">+' + (player.cardCount - 10) + '</span>');
            }
        }
    }

    function updateStatus(response) {
        var $status = $('#status-box');
        if (response.gameEndFlag) {
            $status.text('');
            return;
        }
        var msg = '';
        if (response.hasDrawn) {
            var drawName = playerName(response.lastDrawPlayerIdx);
            var fromName = playerName(response.lastDrawFromIdx);
            msg = drawName + 'が' + fromName + 'から1枚引きました';
            if (response.lastDiscardedPairs > 0) {
                msg += '。' + response.lastDiscardedPairs + '組捨てました';
            }
            msg += '\n';
        }
        // Turn hint
        if (response.currentTurn === 0) {
            var targetName = playerName(response.nextDrawTargetIdx);
            msg += 'あなたの番！ ' + targetName + 'から引いてください。';
        }
        $status.text(msg);
    }

    function updateResult(response) {
        var $result = $('#result-box');
        if (response.message && response.message !== '') {
            $result.text(response.message);
        } else {
            $result.text('');
        }
    }

    function updateButtons(response) {
        if (response.gameEndFlag) {
            $('#btn-draw').prop('disabled', true);
            $('#btn-reset').prop('disabled', false);
        } else {
            var isHumanTurn = (response.currentTurn === 0);
            $('#btn-draw').prop('disabled', !isHumanTurn);
            $('#btn-reset').prop('disabled', false);
        }
    }

    // --- ヘルパー ---

    function playerName(idx) {
        if (idx === 0) return 'あなた';
        return 'CPU ' + idx;
    }

    function cardLabel(card) {
        if (card.design === 'JOKER') return 'JOKER';
        return card.design + ' ' + card.value;
    }

    function getImagePath(card) {
        var prefix;
        if (card.design === 'SPADE') prefix = 's';
        else if (card.design === 'CLOVER') prefix = 'c';
        else if (card.design === 'HEART') prefix = 'h';
        else if (card.design === 'DIAMOND') prefix = 'd';
        else prefix = 'x'; // JOKER
        return './images/' + prefix + zeroPadding(card.value, 2) + '.png';
    }

    function zeroPadding(num, length) {
        return ('0000000000' + num).slice(-length);
    }

    // --- 初期化 ---
    $('#btn-reset').click();
});
