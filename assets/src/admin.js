import { Application } from '@hotwired/stimulus';
import { IdleTimer } from './utils/IdleTimer';

const app = Application.start();
const ctx = require.context('./admin', true, /\.js$/);

// Global variables
// window.userIdle = false; // Track user activity
// window.pushNotification = null; // Push notification object
// window.hidePrompt = null; // Function to hide the prompt
// window.showPrompt = null; // Function to show the prompt

// Idle timer
const idle = new IdleTimer(1 * 60 * 1000, () => {
    window.userIdle = true; // Set the user as idle
}, () => {
    window.userIdle = false; // Set the user as active again
});

ctx.keys().forEach((key) => {
    app.register(
        key.replace(/^\.\//, '').replace(/\.js$/, '').replace(/_controller/, ''),
        ctx(key).default
    );
});