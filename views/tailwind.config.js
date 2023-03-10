/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ['*.{html,js}'],
    theme: {
        extend: {},
    },
    daisyui: {
        themes: ['lofi'],
    },
    plugins: [require('daisyui')],
};
