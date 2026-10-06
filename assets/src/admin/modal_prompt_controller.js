import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    static targets = [
        'title', 'message', 'accept', 'reject'
    ]

    connect() {
        const hidePrompt = () => {
            this.element.classList.remove('is-active');
        };

        const showPrompt = (prompt) => {
            this.titleTarget.textContent      = prompt.title   || "Confirm Action";
            this.messageTarget.textContent    = prompt.message || "Accept of Reject?";
            this.acceptTarget.textContent     = prompt.accept  ||  "Yes";
            this.rejectTarget.textContent     = prompt.reject  ||  "No";
            this.element.classList.add('is-active');

            return new Promise((accept, reject) => {                
                let rejectors = [this.rejectTarget, ...document.querySelectorAll('.modal-background, .modal-close, .modal-card-head .delete') || []];
                rejectors.forEach((close) => {
                    close.addEventListener('click', () => {
                        hidePrompt(); reject();
                    }, {once : true});
                });

                this.acceptTarget.addEventListener('click', () => {
                    hidePrompt(); accept();
                }, {once : true});
            })
        };

        window.prompt = {
            hide: hidePrompt.bind(this),
            show: showPrompt.bind(this)
        };
    }
}