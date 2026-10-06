export class IdleTimer {
    /**
     * @param {number} timeoutMs   How long (ms) before we consider the user idle.
     * @param {() => void} onIdle  Callback to run once when timeout is reached.
     */
    constructor(timeoutMs, onIdle, onReset = null) {
        this.timeoutMs = timeoutMs;
        this.onIdle = onIdle;
        this.onReset = onReset;
        this._timerId = null;
        this._boundReset = this.reset.bind(this);
        this.start();
        console.log('IdleTimer initialized:', timeoutMs, 'ms');
    }

    // Reset or start the inactivity timer
    reset() {
        if (this._timerId) 
            clearTimeout(this._timerId);

        if (this.onReset) {
            this.onReset();
        }

        this._timerId = setTimeout(() => {
            this.onIdle();
        }, this.timeoutMs);
    }

    // Begin tracking
    start() {
        // Watch these events as “activity”
        const events = ['mousemove', 'mousedown', 'keypress', 'touchstart', 'scroll'];
        for (const evt of events) {
            window.addEventListener(evt, this._boundReset, true);
        }
        this.reset();
    }

    // Stop tracking and remove listeners
    stop() {
        clearTimeout(this._timerId);
        window.removeEventListener('mousemove', this._boundReset, true);
        window.removeEventListener('mousedown', this._boundReset, true);
        window.removeEventListener('keypress', this._boundReset, true);
        window.removeEventListener('touchstart', this._boundReset, true);
        window.removeEventListener('scroll', this._boundReset, true);
    }
}