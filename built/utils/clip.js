import nullOrEmpty from './null-or-empty.js';
export default function (s, max) {
    if (nullOrEmpty(s)) {
        return s;
    }
    s = s.trim();
    if (s.length > max) {
        return s.substr(0, max) + '...';
    }
    else {
        return s;
    }
}
