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
        window.pushNotification = this.NotificationsPush.bind(this);
        setInterval(
            this.Notifications.bind(this), 
            100
        );
        
        this.SSEventListen();
    }

    Notifications () {
        this.messages.sort((a, b) => {
            return (b.priority || 0) - (a.priority || 0);
        });
        if(this.messages.length > 0 && this.element.classList.contains('is-hidden')) {
            let message = this.messages.shift();
           
            this.messageTarget.innerHTML = message.message;
            
            let oldClassName = this.element.dataset.type;
            let newClassName = 'is-' + message.type;
            this.element.dataset.type = newClassName;
            
            this.element.classList.remove('is-hidden');
            this.element.classList.remove(oldClassName);
            this.element.classList.add(newClassName);
        }
        this.pendingTarget.innerHTML = Math.max(this.messages.length, 0);
    }

    NotificationsPush(message)
    {
        this.messages.push(message);
    }

    SSEventListen () {
        let eventSource = new EventSource("/admin/ssevents");
        eventSource.onmessage = (event) => {
            let data = JSON.parse(event.data)
            this.NotificationsPush({"type": data.severity, "message":data.message});
        }
    }
}