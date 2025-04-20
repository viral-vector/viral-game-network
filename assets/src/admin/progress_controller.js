import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    static targets = ["progress"]

    connect() {
        window.progress = {
            show: this.show.bind(this),
            hide: this.hide.bind(this),
        };


        this.hide();

        // const lerp = (x, y, a) => x * (1 - a) + y * a;
        setInterval(() => {
            if (this.ishow) {
                this.progressTarget.style.display = "block"
            } else {
                this.progressTarget.style.display = "none"
            }
            // this.progressTarget.value = lerp(this.progressTarget.value, this.value, 0.1)
            // if (this.progressTarget.value >= 99) {
            //     this.stop()
            // }
        }, 100);
    }

    show() {
        console.log("show")
        this.ishow = true
        // this.progressTarget.style.display = "block"
    }

    hide() {
        this.ishow = false
        // this.progressTarget.style.display = "none"
    }
    
    set(val) {
        this.ishow = true
        this.value = Math.max(0, Math.min(100, val))
    }
}