import './style.scss';

addEventListener('load', () => {
    document.getElementById('registerForm').addEventListener('submit', (ev) => {
        ev.preventDefault();

        const userName = (document.getElementById('name') as HTMLInputElement).value;
        const password = (document.getElementById('password') as HTMLInputElement).value;
        const passwordConfirm = (document.getElementById('passwordConfirm') as HTMLInputElement)
            .value;

        const errorBar = document.getElementById('errorBar');
        if (password !== passwordConfirm) {
            errorBar.innerText = 'パスワードが一致しません';
            errorBar.setAttribute('data-activated', 'yes');
            return;
        }
        errorBar.setAttribute('data-activated', 'no');

        fetch('/users/new', {
            method: 'POST',
            body: `userName=${encodeURI(userName)}&password=${encodeURI(password)}`,
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
        })
            .then((resp) => resp.json())
            .then((resp) => {
                if (!resp['success']) {
                    errorBar.innerText = resp['reason'];
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
