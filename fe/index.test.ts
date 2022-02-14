import * as index from './index';

test('add', () => {
    expect(index.plus(1, 2)).toBe(3);
    expect(index.plus(1, -2)).toBe(-1);
});
