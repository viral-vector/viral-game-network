import { Application } from '@hotwired/stimulus';

const app = Application.start();
const ctx = require.context('./admin', true, /\.js$/);

ctx.keys().forEach((key) => {
    app.register(
        key.replace(/^\.\//, '').replace(/\.js$/, '').replace(/_controller/, ''),
        ctx(key).default
    );
});