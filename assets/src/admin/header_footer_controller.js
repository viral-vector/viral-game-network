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
            let refs = [];
            const hsetAttr = item.getAttribute("hset");
            if (hsetAttr) {
                refs = hsetAttr.split(",").map(r => r.trim()).filter(r => r !== "");
            }
            refs.push(item.getAttribute("href"));
            const active = refs.some(ref => {
                if (this.routeValue in kmap && kmap[this.routeValue].absolute) {
                return ref === this.routeValue;
                }
                return ref === this.routeValue || ref.startsWith(this.routeValue) || this.routeValue.startsWith(ref);
            });

            if (active) 
                item.classList.add("is-active");
        }
        [...document.querySelectorAll('.navbar-burger')].forEach(el => {
            el.addEventListener('click', () => {
                el.classList.toggle('is-active');
                document.getElementById(el.dataset.target).classList.toggle('is-active');
            });
        });
    }
}