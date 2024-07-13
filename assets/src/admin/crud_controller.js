import { Controller } from "@hotwired/stimulus"

export default class extends Controller {

    static values = {
        create: String,
        delete: String,
        list: String,
        view: String,
    }

    connect() {
    }

    async delete(event) {
        let execute = true;
        try {
            let prompt = {
                title: "Delete Item",
                message: "Are you sure you want to delete this item?",
                accept: "Yes",
                reject: "No",
            };
            await window.showPrompt(prompt);
        } catch (error) {
            execute = false;
        }
        if (!execute) {
            return;
        }

        let url = this.deleteValue.replace(":id", event.target.dataset.id);
        fetch(url, {
            method: "DELETE",
        }).then(response => {
            if (response.ok) {
                
            }
            // Remove Item or Redirect to List
        });
    }

    async create(event) {
        let execute = true;
        try {
            let prompt = {
                title: "Create Item",
                message: "Are you sure you want to create a new item?",
                accept: "Yes",
                reject: "No",
            };
            await window.showPrompt(prompt);
        } catch (error) {
            execute = false;
        }
        if (!execute) {
            return;
        }

        let url = this.createValue;
        fetch(url, {
            method: "POST",
        }).then(response => {
            if (response.ok) {
                
            }
            // Add to list or Redirect to item
        });
    }
}