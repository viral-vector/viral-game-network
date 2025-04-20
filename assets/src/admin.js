import './pwa_manifest.json';
import './pwa_service_worker.js';

import { Application } from '@hotwired/stimulus';
import { IdleTimer } from './utils/IdleTimer'; 

const app = Application.start();
const ctx = require.context('./admin', true, /\.js$/);

// Global variables
// window.userIdle = false; // Track user activity
// window.pushNotification = null; // Push notification object
// window.prompt = {}; // Prompt object
// window.progress = {}; // Progress object

// Idle timer
const idle = new IdleTimer(1 * 60 * 1000, () => {
    window.userIdle = true; // Set the user as idle
}, () => {
    window.userIdle = false; // Set the user as active again
});

// Service worker registration
if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/service_worker.js')
    .then(registration => {
        console.log('Service Worker registered with scope:', registration.scope);
    })
    .catch(error => {
        console.error(error);
    });
}

// Automatically register Stimulus controllers
ctx.keys().forEach((key) => {
    app.register(
        key.replace(/^\.\//, '').replace(/\.js$/, '').replace(/_controller/, ''),
        ctx(key).default
    );
});