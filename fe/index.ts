'use strict';
import './style.scss';

addEventListener('devicemotion', function (ev) {
    if (ev === null || ev.acceleration === null) {
        return;
    }
    const ax = ev.acceleration.x || 0;
    document.body.innerHTML = ax.toString();
});
