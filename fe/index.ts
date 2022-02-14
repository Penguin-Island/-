'use strict';
import './style.scss';

export function plus(a, b) {
    return a + b;
}

const a: string = 'fuga1';
document.addEventListener('DOMContentLoaded', () => {
    console.log(plus(1, 2));
});
