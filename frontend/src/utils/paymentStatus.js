export const normalizePaymentStatus = (status = '') => {
    const normalized = status.toLowerCase();

    if (["pending", "pending_payment", "processing", "authorized"].includes(normalized)) {
        return "pending";
    }

    if (["completed", "paid", "approved", "succeeded"].includes(normalized)) {
        return "completed";
    }

    if (["failed", "rejected", "cancelled", "canceled"].includes(normalized)) {
        return "failed";
    }

    if (["refunded", "chargeback"].includes(normalized)) {
        return "refunded";
    }

    return normalized || "unknown";
};
