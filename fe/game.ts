import './game.scss';

//TODO: remove this after implementing login.
fetch('/testing', {
    method: 'post',
});

const bgm = new Audio('/assets/ryugen.mp3');
bgm.loop = true;
const seStart = new Audio('/assets/start.mp3');
const seTurnChange = new Audio('/assets/turn.mp3');
const seAlarm = new Audio('/assets/alarm.mp3');
seAlarm.loop = true;

const playAndPause = (audio) => {
    audio.play();
    audio.pause();
};

const showUserInfo = () => {
    fetch('/api/user/info')
        .then((resp) => resp.json())
        .then((resp) => {
            document.getElementById('playerName').innerText = resp['userName'];
            document.getElementById('playerId').innerText = resp['playerTag'];

            document.getElementById('successRate').innerText = resp['successRate'];
            if (resp['joinedGroup']) {
                document.getElementById('startTime').innerText = resp['groupInfo']['wakeUpTime'];
                document.getElementById('timeContainer').setAttribute('data-activated', 'yes');

                const friendsContainer = document.getElementById('friends');
                if (resp['groupInfo']['members'].length > 0) {
                    friendsContainer.innerHTML = '';
                    for (const friend of resp['groupInfo']['members']) {
                        const item = document.createElement('div');
                        item.classList.add('friend-name');
                        item.innerText = friend;
                        friendsContainer.appendChild(item);
                    }
                }
            } else {
                document.getElementById('noFriendsTip').setAttribute('data-activated', 'yes');
                document
                    .getElementById('startButtonContainer')
                    .setAttribute('data-activated', 'no');
            }
        })
        .catch((err) => {
            console.error(err);
            document.getElementById('alertMessage').innerText = 'ユーザー情報の取得に失敗しました';
            document.getElementById('alert').setAttribute('data-activated', 'yes');
        });
};

addEventListener('load', () => {
    showUserInfo();

    let sock = null;
    let stillWaitingRetry = false;

    document.getElementById('startGame').addEventListener('click', (ev) => {
        const target = ev.target as HTMLInputElement;
        target.innerText = '相手を待っています…';
        target.disabled = true;

        playAndPause(bgm);
        playAndPause(seStart);
        playAndPause(seTurnChange);
        playAndPause(seAlarm);

        sock = new WebSocket('ws://localhost:8000/api/ws');
        sock.addEventListener('close', () => {
            console.log('close');
        });
        sock.addEventListener('error', () => {
            console.log('error');
        });
        sock.addEventListener('message', (ev) => {
            const data = JSON.parse(ev.data);
            if (data['type'] == 'onStart') {
                document.getElementById('top').setAttribute('data-activated', 'no');
                document.getElementById('game').setAttribute('data-activated', 'yes');

                seStart.play();
                bgm.play();
            } else if (data['type'] == 'onChangeTurn') {
                document.getElementById('prevWord').innerText = data['data']['prevAnswer'];

                const yourTurn = data['data']['yourTurn'];
                document.getElementById('turn').innerText = yourTurn ? 'あなたの番' : '相手の番';
                (document.getElementById('send') as HTMLInputElement).disabled = !yourTurn;
                (document.getElementById('wordInput') as HTMLInputElement).disabled = !yourTurn;

                if (yourTurn) {
                    seTurnChange.play();
                }
            } else if (data['type'] == 'onTick') {
                document.getElementById('countDown').innerText = data['data']['remainSec'];

                if (data['data']['waitingRetry']) {
                    if (!stillWaitingRetry) {
                        bgm.pause();
                        seAlarm.play();
                        stillWaitingRetry = true;
                    }

                    document.getElementById('failureOverlay').setAttribute('data-activated', 'yes');
                    document.getElementById('finishOverlay').setAttribute('data-activated', 'no');

                    (document.getElementById('confirmRetry') as HTMLInputElement).disabled = false;
                } else if (data['data']['finished']) {
                    bgm.pause();
                    seAlarm.pause();
                    setTimeout(() => {
                        location.href = '/finish/';
                    }, 3000);

                    document.getElementById('failureOverlay').setAttribute('data-activated', 'no');
                    document.getElementById('finishOverlay').setAttribute('data-activated', 'yes');
                } else {
                    if (stillWaitingRetry) {
                        stillWaitingRetry = false;
                        bgm.play();
                        seAlarm.pause();
                    }
                    document.getElementById('failureOverlay').setAttribute('data-activated', 'no');
                    document.getElementById('finishOverlay').setAttribute('data-activated', 'no');
                    document.getElementById('turnCountDown').innerText =
                        data['data']['turnRemainSec'];
                }
            } else if (data['type'] == 'onError') {
                document.getElementById('alertMessage').innerText = data['data']['reason'];
                document.getElementById('alert').setAttribute('data-activated', 'yes');
            }
        });
    });

    document.getElementById('closeAlert').addEventListener('click', () => {
        document.getElementById('alert').setAttribute('data-activated', 'no');
    });

    document.getElementById('wordInput').addEventListener('input', (event) => {
        const ev = event as InputEvent;
        if (ev.isComposing) {
            return;
        }
        const inputElement = ev.target as HTMLInputElement;
        const str = inputElement.value;
        const resultStr = [];
        for (let i = 0; i < str.length; i++) {
            const cc = str.charCodeAt(i);
            if ((0x3041 <= cc && cc <= 0x3096) || cc == 0x30fc) {
                resultStr.push(cc);
            } else if (0x30a1 <= cc && cc <= 0x30f6) {
                resultStr.push(cc - 96);
            }
        }
        inputElement.value = String.fromCharCode(...resultStr);
    });

    document.getElementById('send').addEventListener('click', (ev) => {
        ev.preventDefault();
        if (sock === null) {
            return;
        }

        const input = document.getElementById('wordInput') as HTMLInputElement;
        sock.send(
            JSON.stringify({
                type: 'sendAnswer',
                data: {word: input.value},
            })
        );
        input.value = '';
    });

    document.getElementById('confirmRetry').addEventListener('click', (ev) => {
        ev.preventDefault();
        if (sock === null) {
            return;
        }

        (ev.target as HTMLInputElement).disabled = true;

        sock.send(JSON.stringify({type: 'confirmRetry', data: {}}));
    });
});
