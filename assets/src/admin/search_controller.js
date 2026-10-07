import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    static targets = ['control', 'button']
    static values = {
        action: String
    }
    connect() {
        this.disconnect();
        this.button = this.buttonTarget;
        this.onSearch = (event) => {
            event.preventDefault();
            const search = this.controlTarget.value.trim();
            if (search.length > 0 && search.length < 3) {
                window.pushNotification?.({
                    "type": "warning",
                    "message": "Please enter at least 3 characters to search.",
                    "priority": 0
                })
                return
            }
            if (this.navigationTimer !== undefined) return;
            this.button.disabled = true;
            this.navigationTimer = setTimeout(() => {
                this.navigationTimer = undefined;
                const url = new URL(this.actionValue || window.location.href, window.location.href);
                if (search) url.searchParams.set('search', search);
                else url.searchParams.delete('search');
                url.searchParams.set('page', 1);
                url.searchParams.delete('pageToken');
                url.searchParams.delete('prevPage');
                window.location.href = url.toString();
            }, 100);
        };
        this.button.addEventListener('click', this.onSearch);
    }

    disconnect() {
        clearTimeout(this.navigationTimer);
        this.navigationTimer = undefined;
        if (this.button && this.onSearch) {
            this.button.removeEventListener('click', this.onSearch);
            this.button.disabled = false;
        }
    }
}
