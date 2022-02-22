import './style.scss';

addEventListener('load', () => {
    document.getElementById('registerForm').addEventListener('submit', (ev) => {
        ev.preventDefault();

        const userName = (document.getElementById('name') as HTMLInputElement).value;
        const password = (document.getElementById('password') as HTMLInputElement).value;

        fetch('/users/new', {
            method: 'POST',
            body: `username=${encodeURI(userName)}&password=${encodeURI(password)}`,
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
        })
            .then((resp) => resp.json())
            .then((resp) => {
                if (!resp['success']) {
                    const errorBar = document.getElementById('errorBar');
                    errorBar.innerText = resp['reason'];
                    errorBar.setAttribute('data-activated', 'yes');
                    return;
                }
                location.href = '/game/';
            })
            .catch((err) => {
                const errorBar = document.getElementById('errorBar');
                errorBar.innerText = 'サーバーとの通信に失敗しました';
                errorBar.setAttribute('data-activated', 'yes');
            });
    });
});
