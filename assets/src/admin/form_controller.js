import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    connect() {
        console.log('Form Controller Connected!')
        this.buildAForms()
    }

    buildAForms() {
        this.element.querySelectorAll('form').forEach(element => {
            element.querySelectorAll('input[is-checked="true"]').forEach(item => {
                item.click()
                item.setAttribute('checked', true)
            })
            element.querySelectorAll('select').forEach(item => {
                for (let opt of item.querySelectorAll('option')) {
                    if (opt.getAttribute('value') == item.getAttribute('value')) {
                        opt.click()
                        opt.setAttribute('selected', true)
                    }
                }
            })
            element.querySelectorAll('input[type="text"][is-disabled="true"]').forEach(item => {
                item.setAttribute('readonly', true)
            })
            element.querySelectorAll('input[type="checkbox"][is-disabled="true"]').forEach(item => {
                let inp = document.createElement("input");
                inp.setAttribute('type', 'hidden')
                inp.setAttribute('name', item.getAttribute('name'))
                inp.setAttribute('value', item.getAttribute('value'))

                item.parentNode.appendChild(inp)
                item.setAttribute('disabled', true)
                item.removeAttribute('name')
            })

            element.addEventListener('submit', async event => {
                event.preventDefault();

                let execute = true;
                if (event.target.dataset.confirmMessage) {
                    try {
                        let prompt = {
                            title: event.target.dataset.confirmTitle,
                            message: event.target.dataset.confirmMessage,
                            accept: event.target.dataset.confirmAccept,
                            reject: event.target.dataset.confirmReject
                        };
                        await window.showPrompt(prompt);
                    } catch (error) {
                        execute = false;
                    }
                }
                if (!execute) {
                    return;
                }

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
                    window.pushNotification({
                        "type": "info",
                        "message": data.message,
                        "priority": 0
                    })
                } catch (error) {
                    console.error(error.message);
                    // Show an error toast
                    window.pushNotification({
                        "type": "danger",
                        "message": error.message,
                        "priority": 1
                    })
                }
            })
        })
    }
}