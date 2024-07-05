import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    static values = {
        route: String
    }

    connect() {
        for (const item of this.element.querySelectorAll("a.menu-item")) {
            if (item.getAttribute("href") != this.routeValue) {
                continue;
            }
            item.classList.add("is-active");
        }

        [...document.querySelectorAll('.navbar-burger')].forEach(el => {
            el.addEventListener('click', () => {
                el.classList.toggle('is-active');
                document.getElementById(el.dataset.target).classList.toggle('is-active');
            });
        });

        for (const item of this.element.querySelectorAll(".navbar-item.has-dropdown")) {
            item.addEventListener("click", (event) => {
                event.preventDefault();
                item.classList.toggle("is-active");
            });
        }
    }
}