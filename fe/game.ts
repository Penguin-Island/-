import './game.scss';

//TODO: remove this after implementing login.
fetch('/testing', {
    method: 'post',
});

addEventListener('load', () => {
    const sock = new WebSocket('ws://localhost:8000/api/ws');
    sock.addEventListener('close', () => {
        console.log('close');
    });
    sock.addEventListener('error', () => {
        console.log('error');
    });
    sock.addEventListener('message', (ev) => {
        const data = JSON.parse(ev.data);
        if (data['type'] == 'onChangeTurn') {
            document.getElementById('prevWord').innerText = data['data']['prevAnswer'];

            const yourTurn = data['data']['yourTurn'];
            document.getElementById('turn').innerText = yourTurn ? 'あなたの番' : '相手の番';
            (document.getElementById('send') as HTMLInputElement).disabled = !yourTurn;
            (document.getElementById('wordInput') as HTMLInputElement).disabled = !yourTurn;
        } else if (data['type'] == 'onTick') {
            document.getElementById('countDown').innerText = data['data']['remainSec'];

            if (data['data']['waitingRetry']) {
                document.getElementById('failureOverlay').setAttribute('data-activated', 'yes');
                document.getElementById('finishOverlay').setAttribute('data-activated', 'no');

                (document.getElementById('confirmRetry') as HTMLInputElement).disabled = false;
            } else if (data['data']['finished']) {
                document.getElementById('failureOverlay').setAttribute('data-activated', 'no');
                document.getElementById('finishOverlay').setAttribute('data-activated', 'yes');
            } else {
                document.getElementById('failureOverlay').setAttribute('data-activated', 'no');
                document.getElementById('finishOverlay').setAttribute('data-activated', 'no');
                document.getElementById('turnCountDown').innerText = data['data']['turnRemainSec'];
            }
        }
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
        (ev.target as HTMLInputElement).disabled = true;

        sock.send(JSON.stringify({type: 'confirmRetry', data: {}}));
    });
});
