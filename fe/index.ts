import './style.scss';

addEventListener('load', () => {
    document.getElementById('loginForm').addEventListener('submit', (ev) => {
        ev.preventDefault();

        const userName = (document.getElementById('userName') as HTMLInputElement).value;
        const password = (document.getElementById('password') as HTMLInputElement).value;

        const errorBar = document.getElementById('errorBar');

        fetch('/users/login', {
            method: 'post',
            body: `userName=${encodeURI(userName)}&password=${encodeURI(password)}`,
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
        })
            .then((resp) => {
                if (resp.status !== 200) {
                    errorBar.innerText = 'ユーザー名またはパスワードが間違っています';
                    errorBar.setAttribute('data-activated', 'yes');
                    return;
                }
                location.href = '/game/';
            })
            .catch((err) => {
                errorBar.innerText = 'サーバーとの通信に失敗しました';
                errorBar.setAttribute('data-activated', 'yes');
            });
    });
});
