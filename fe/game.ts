import './style.scss';

fetch('/testing', {
    method: 'post',
});

const sock = new WebSocket('ws://localhost:8000/api/ws');
sock.addEventListener('open', () => {
    console.log('open');
});
sock.addEventListener('close', () => {
    console.log('close');
});
sock.addEventListener('error', () => {
    console.log('error');
});
sock.addEventListener('message', (ev) => {
    console.log('message: ', ev.data);
    const data = JSON.parse(ev.data);
    if (data['type'] == 'onChangeTurn') {
        document.getElementById('prevWord').innerText = data['data']['prevAnswer'];
        document.getElementById('yourTurn').innerText = data['data']['yourTurn'];
    } else if (data['type'] == 'onTick') {
        document.getElementById('remain').innerText = data['data']['remainSec'];
        document.getElementById('turnRemain').innerText = data['data']['turnRemainSec'];
        document.getElementById('waitingRetry').innerText = data['data']['waitingRetry'];
        document.getElementById('finished').innerText = data['data']['finished'];
    }
});

document.getElementById('send').addEventListener('click', () => {
    sock.send(
        JSON.stringify({
            type: 'sendAnswer',
            data: {word: (document.getElementById('wordInput') as HTMLInputElement).value},
        })
    );
});

document.getElementById('confirmRetry').addEventListener('click', () => {
    sock.send(JSON.stringify({type: 'confirmRetry', data: {}}));
});
