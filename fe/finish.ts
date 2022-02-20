import './finish.scss';

addEventListener('load', () => {
    fetch('/users/statistics')
        .then((resp) => resp.json())
        .then((data) => {
            if (data === 0) {
                const errBar = document.getElementById('errorBar');
                errBar.innerText = 'データがありません';
                errBar.setAttribute('data-has-error', 'yes');
                return;
            }
            data.reverse();

            document.getElementById('month').innerText = data[0]['month'].toString();
            document.getElementById('day').innerText = data[0]['day'].toString();
            document
                .getElementById('titleContainer')
                .setAttribute('data-success', data[0]['success'] ? 'yes' : 'no');

            const container = document.getElementById('resultContainer');
            const rowTemplate = document.getElementById('resultRow') as HTMLTemplateElement;
            for (const e of data.slice(1)) {
                const row = (rowTemplate.content.cloneNode(true) as DocumentFragment)
                    .firstElementChild;
                container.appendChild(row);
                (
                    row.querySelector('.date') as HTMLElement
                ).innerText = `${e['month']}月${e['day']}日`;
                const resultElement = row.querySelector('.result') as HTMLElement;
                resultElement.innerText = e['success'] ? '〇' : '✘';
                resultElement.classList.add(e['success'] ? 'success' : 'failure');
            }
        })
        .catch((err) => {
            console.error(err);
            document.getElementById('errorBar').setAttribute('data-has-error', 'yes');
        });
});
