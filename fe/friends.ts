import './friends.scss';

addEventListener('load', () => {
    let playerTag = null;

    document.getElementById('searchButton').addEventListener('click', (ev) => {
        ev.preventDefault();

        playerTag = (document.getElementById('playerTag') as HTMLInputElement).value;
        fetch(`/users/find?playerTag=${encodeURI(playerTag)}`)
            .then((resp) => {
                if (resp.status != 302) {
                    document.getElementById('message').innerText = '見つかりません';
                    return;
                }
                document.getElementById('userName').innerText = playerTag;
                document.getElementById('resultContainer').setAttribute('data-found', 'yes');
            })
            .catch((err) => {
                document.getElementById('resultContainer').setAttribute('data-found', 'no');
            });
    });

    document.getElementById('inviteButton').addEventListener('click', () => {
        fetch('/groups/invite', {
            method: 'post',
            body: `player=${encodeURI(playerTag)}`,
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
        })
            .then((resp) => {
                if (resp.status !== 201) {
                    document.getElementById('resultContainer').setAttribute('data-found', 'no');
                    document.getElementById('message').innerText = '招待に失敗しました';
                    return;
                }
                document.getElementById('resultContainer').setAttribute('data-found', 'no');
                (document.getElementById('playerTag') as HTMLInputElement).value = '';
                document.getElementById('message').innerText = '招待しました';
            })
            .catch((err) => {
                document.getElementById('resultContainer').setAttribute('data-found', 'no');
                document.getElementById('message').innerText = '招待に失敗しました';
            });
    });
});
