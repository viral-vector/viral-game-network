import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    static targets = [
        'pending', 'message'
    ];

    connect() {
        this.messages = [
            // {"type":"info", "message":"Hello World 1", "priority": 1},
        ];

        (this.element.querySelectorAll('.delete') || []).forEach(elem => {
            elem.addEventListener('click', () => {
               this.element.classList.add('is-hidden');
            });
        });
        this.pushNotification = this.NotificationsPush.bind(this);
        window.pushNotification = this.pushNotification;
        this.notificationTimer = setInterval(
            () => this.Notifications(),
            100
        );

        this.SSEventListen();
    }

    disconnect() {
        clearInterval(this.notificationTimer);
        if (this.eventSource) {
            this.eventSource.onmessage = null;
            this.eventSource.close();
        }
        if (window.pushNotification === this.pushNotification) delete window.pushNotification;
    }

    Notifications () {
        this.messages.sort((a, b) => {
            return (b.priority || 0) - (a.priority || 0);
        });
        if(this.messages.length > 0 && this.element.classList.contains('is-hidden')) {
            let message = this.messages.shift();

            this.messageTarget.textContent = message.message;

            let oldClassName = this.element.dataset.type;
            const type = ['info', 'success', 'warning', 'danger', 'primary', 'link'].includes(message.type) ? message.type : 'info';
            let newClassName = 'is-' + type;
            this.element.dataset.type = newClassName;

            this.element.classList.remove('is-hidden');
            if (oldClassName) this.element.classList.remove(oldClassName);
            this.element.classList.add(newClassName);
        }
        this.pendingTarget.textContent = Math.max(this.messages.length, 0);
    }

    NotificationsPush(message)
    {
        if (message && typeof message.message === "string") this.messages.push(message);
    }

    SSEventListen () {
        this.eventSource = new EventSource("/admin/ssevents");
        this.eventSource.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                if (data) this.NotificationsPush({ type: data.severity, message: data.message });
            } catch {
                // A malformed event must not stop the remaining notification stream.
            }
        };
    }
}
