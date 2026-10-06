import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    connect() {
        // Check if the token is expiring every minute
        this.refreshInterval = setInterval(() => {
            if (this.isJWTTokenExpire()) {
                this.getFreshJWTToken();
            }
        }, 1000 * 45);
    }

    disconnect() {
        clearInterval(this.refreshInterval);
    }

    isJWTTokenExpire() {
        const token = this.getCookie('VNET_SESSION');
        if (!token) return false;
        try {
            const payload = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/');
            const { exp } = JSON.parse(atob(payload));
            const difference = exp * 1000 - Date.now();
            return Number.isFinite(exp) && 0 < difference && difference < 5 * 60 * 1000;
        } catch {
            return false;
        }
    }

    async getFreshJWTToken() {
        if (document.hidden || window.userIdle) {
            return;
        }
        const response = await fetch('/auth/admin/refresh', {
            method: 'GET',
        });
        if (response.status === 200) {
            const data = await response.json();
            if (data && data.access_token) {
                const tokenData = JSON.parse(atob(data.access_token.split('.')[1]));
                this.setCookie("VNET_SESSION", data.access_token, {
                    expires: new Date(tokenData.exp * 1000),
                    secure: true,
                    sameSite: "Strict"
                });
                console.log('Token refreshed successfully');
            }
        }
    }

    getCookie(name) {
        // escape any special regex chars in the cookie name
        const escaped = name.replace(/([.*+?^${}()|[\]\\])/g, '\\$1');
        //   (^|;\s*)   … start of string or “; ”  
        //   name=      … literal cookie name + “=”  
        //   ([^;]*)    … capture everything up to next semicolon (the value)
        const match = document.cookie.match(new RegExp('(?:^|;\\s*)' + escaped + '=([^;]*)'));
        // If not found, match is null
        return match ? decodeURIComponent(match[1]) : undefined;
    }

    /**
     * Set a cookie with the given name, value, and options.
     *
     * @param {string} name  - The cookie name.
     * @param {string} value - The cookie value.
     * @param {Object} [options] - Optional cookie attributes.
     *   @property {number|Date} [expires]  - Days from now or a Date object.
     *   @property {string}           [path]     - URL path (default "/").
     *   @property {string}           [domain]   - Cookie domain.
     *   @property {boolean}          [secure]   - Secure flag.
     *   @property {"Lax"|"Strict"|"None"} [sameSite] - SameSite policy.
     */
    setCookie(name, value, options = {}) {
        // Encode name and value to escape semicolons, etc.
        let cookie = encodeURIComponent(name) + "=" + encodeURIComponent(value);

        // Handle expires option: number of days or a Date object
        if (options.expires) {
            let expires = options.expires;
            if (typeof expires === "number") {
                const date = new Date();
                date.setTime(date.getTime() + expires * 24 * 60 * 60 * 1000);
                expires = date;
            }
            cookie += "; expires=" + expires.toUTCString();
        }

        // Append other attributes if provided
        if (options.path) cookie += "; path=" + options.path;
        if (options.domain) cookie += "; domain=" + options.domain;
        if (options.secure) cookie += "; secure";
        if (options.sameSite) cookie += "; samesite=" + options.sameSite;

        // Write the cookie
        document.cookie = cookie;
    }
}