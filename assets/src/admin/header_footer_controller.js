import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    static values = {
        route: String
    }

    connect() {
        const menu = this.element.querySelectorAll("a.menu-item");
        const kmap = {};
        for (const item of menu) {
            let href = item.getAttribute("href");
            kmap[href] = {
                "absolute": item.getAttribute("is-absolute")
            }
        }
        for (const item of menu) {
            let href = item.getAttribute("href");
            let active = href.startsWith(this.routeValue);
            if(this.routeValue in kmap && kmap[this.routeValue].absolute){
                active = href === this.routeValue;
            }
            if (active)
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