import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    static targets = [
        'title', 'message', 'accept', 'reject'
    ]

    connect() {
        const closePrompt = () => {
            this.element.classList.remove('is-active');
        }

        window.hidePrompt = closePrompt;
        window.showPrompt = (prompt) => {
            this.titleTarget.innerHTML      = prompt.title   || "Confirm Action";
            this.messageTarget.innerHTML    = prompt.message || "Accept of Reject?";
            this.acceptTarget.innerHTML     = prompt.accept  ||  "Yes";
            this.rejectTarget.innerHTML     = prompt.reject  ||  "No";
            this.element.classList.add('is-active');

            return new Promise((accept, reject) => {                
                let rejectors = [this.rejectTarget, ...document.querySelectorAll('.modal-background, .modal-close, .modal-card-head .delete') || []];
                rejectors.forEach((close) => {
                    close.addEventListener('click', () => {
                        closePrompt(); reject();
                    }, {once : true});
                });

                this.acceptTarget.addEventListener('click', () => {
                    closePrompt(); accept();
                }, {once : true});
            })
        }
    }
}