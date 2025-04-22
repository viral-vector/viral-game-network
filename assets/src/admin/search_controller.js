import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    static targets = ['control', 'button']
    static values = {
        action: String
    }
    connect() {
        // if not empty string
        if (this.actionValue === undefined || this.actionValue === '') {
            this.actionValue = window.location.toLocaleString();
        }

        const url = new URL(this.actionValue);
        this.buttonTarget.addEventListener('click', (e) => {
            if (this.controlTarget.value.length > 0 && this.controlTarget.value.length < 3) {
                window.pushNotification({
                    "type": "warning",
                    "message": "Please enter at least 3 characters to search.",
                    "priority": 0
                })
                return
            }
            this.buttonTarget.setAttribute('disabled', 'disabled')
            setTimeout(() => {
                url.searchParams.set('search', this.controlTarget.value)
                window.location.href = url.toString();
            }, 100);
        });
    }
}