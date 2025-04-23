import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    connect() {
        this.buildAForms()
    }

    buildAForms() {
        this.element.querySelectorAll('form').forEach(element => {
            element.querySelectorAll('.vgnform-input[is-checked="true"]').forEach(item => {
                item.click()
                item.setAttribute('checked', true)
            })
            element.querySelectorAll('select.vgnform-input').forEach(item => {
                for (let opt of item.querySelectorAll('option')) {
                    if (opt.getAttribute('value') == item.getAttribute('value')) {
                        opt.click()
                        opt.setAttribute('selected', true)
                    }
                }
            })
            element.querySelectorAll('.vgnform-input[type="text"][is-disabled="true"], .vgnform-input[type="number"][is-disabled="true"]').forEach(item => {
                item.setAttribute('readonly', true)
                item.setAttribute('disabled', true)
                item.removeAttribute('name')
            })
            element.querySelectorAll('.vgnform-input[is-required="true"]').forEach(item => {
                item.setAttribute('required', true)
            })
            element.querySelectorAll('.vgnform-input[type="checkbox"][is-disabled="true"]').forEach(item => {
                let inp = document.createElement("input");
                inp.setAttribute('type', 'hidden')
                inp.setAttribute('name', item.getAttribute('name'))
                inp.setAttribute('value', item.getAttribute('value'))

                item.parentNode.appendChild(inp)
                item.setAttribute('disabled', true)
                item.removeAttribute('name')
            })

            element.addEventListener('submit', async event => {
                await this.submitForm(event);
            })
        })
    }

    async submitForm(event) {
        event.preventDefault();


        if (event.target.getAttribute('action') == "")
            return;

        let execute = true;
        if (event.target.dataset.confirmMessage) {
            try {
                let prompt = {
                    title: event.target.dataset.confirmTitle,
                    message: event.target.dataset.confirmMessage,
                    accept: event.target.dataset.confirmAccept,
                    reject: event.target.dataset.confirmReject
                };
                await window.prompt.show(prompt);
            } catch (error) {
                execute = false;
            }
        }
        if (!execute) {
            return;
        }

        window.progress.show()
        try {
            const action = event.target.getAttribute('action');
            const method = event.target.getAttribute('method');
            const formData = new FormData(event.target);
            const response = await fetch(action, {
                method: method,
                body: formData,
            });

            // Parse response JSON
            const data = await response.json();
            if (!response.ok) {
                data.status = response.status;
                throw new Error(`${response.status}: ${JSON.stringify(data) || ''}`);
            }

            if (data.redirect) {
                setTimeout(() => {
                    window.location.href = data.redirect
                }, 500);
            }

            window.pushNotification({
                "type": "info",
                "message": data.message || "success",
                "priority": 0
            })
            event.target.dispatchEvent(new CustomEvent("form-process", {
                bubbles: true,
                detail: data
            }))

        } catch (error) {
            window.pushNotification({
                "type": "danger",
                "message": error.message,
                "priority": 1
            })
            event.target.dispatchEvent(new CustomEvent("form-process", {
                bubbles: true,
                detail: {
                    'error': error.message,
                }
            }))
        }

        setTimeout(() => {
            window.progress.hide()
        }, 1000);
    }
}