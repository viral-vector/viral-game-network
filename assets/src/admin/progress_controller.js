import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    static targets = ["progress"]

    connect() {
        this.progressApi = {
            show: this.show.bind(this),
            hide: this.hide.bind(this),
        };
        window.progress = this.progressApi;
        this.hide();

        // const lerp = (x, y, a) => x * (1 - a) + y * a;
        this.updateInterval = setInterval(() => {
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

    disconnect() {
        clearInterval(this.updateInterval);
        if (window.progress === this.progressApi) delete window.progress;
    }

    show() {
        this.ishow = true
    }

    hide() {
        this.ishow = false
    }
    
    set(val) {
        this.ishow = true
        this.value = Math.max(0, Math.min(100, val))
    }
}
