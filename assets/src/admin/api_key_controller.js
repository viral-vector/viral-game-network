import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    static targets = [
        'addkey', 'addkeyForm', 'keylist'
    ]
    connect() {
        this.addkeyTarget.addEventListener("click", async (event) => {
            this.addkeyFormTarget.style.display = 'block';
        })

        this.addkeyFormTarget.addEventListener('form-process', (event) => {
            if (!event.detail.error) {
                const tmp = document.createElement('div',);
                tmp.innerHTML = event.detail.render;
                if (tmp.firstChild) {
                    this.keylistTarget.appendChild(tmp.firstChild);
                }
            }
        })

        this.element.querySelectorAll('.list form').forEach((elem) => {
            elem.addEventListener('form-process', (event) => {
                if (!event.detail.error) {
                    setTimeout(() => {
                        elem.closest(".list-item").remove()
                    }, 500)
                }
            })
        });
    }
}