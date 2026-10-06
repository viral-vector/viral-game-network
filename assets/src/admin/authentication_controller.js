import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    static values = { expires: Number }

    connect() {
        this.refreshInterval = setInterval(() => {
            if (this.isJWTTokenExpire()) this.getFreshJWTToken();
        }, 45_000);
    }

    disconnect() {
        clearInterval(this.refreshInterval);
    }

    isJWTTokenExpire() {
        const difference = this.expiresValue * 1000 - Date.now();
        return Number.isFinite(this.expiresValue) && 0 < difference && difference < 5 * 60 * 1000;
    }

    async getFreshJWTToken() {
        if (document.hidden || window.userIdle || this.refreshInFlight) return false;
        this.refreshInFlight = true;
        try {
            const response = await fetch('/auth/admin/refresh', { method: 'GET' });
            if (response.redirected) {
                window.location.assign(response.url);
                return false;
            }
            if (!response.ok) return false;
            const data = await response.json();
            if (!Number.isFinite(data.expires_at) || data.expires_at * 1000 <= Date.now()) return false;
            // The backend renews the HttpOnly cookie. Only expiry metadata is public.
            this.expiresValue = data.expires_at;
            return true;
        } catch {
            return false;
        } finally {
            this.refreshInFlight = false;
        }
    }
}
