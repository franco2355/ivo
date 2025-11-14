const STATUS_GROUPS = {
    pending: ["pending", "pending_payment", "processing", "authorized", "in_process"],
    completed: ["completed", "paid", "approved", "succeeded"],
    failed: ["failed", "rejected", "cancelled", "canceled", "denied"],
    refunded: ["refunded", "chargeback", "returned"],
};

const STATUS_LABELS = {
    pending: "Pendiente",
    completed: "Completado",
    failed: "Fallido",
    refunded: "Reembolsado",
    unknown: "Desconocido",
};

const STATUS_ICONS = {
    pending: "⏳",
    completed: "✓",
    failed: "✗",
    refunded: "↩",
    unknown: "?",
};

const STATUS_BADGE_CLASSES = {
    pending: "estado-pendiente",
    completed: "estado-completado",
    failed: "estado-fallido",
    refunded: "estado-reembolsado",
    unknown: "estado-pendiente",
};

const mapStatus = new Map(
    Object.entries(STATUS_GROUPS).flatMap(([key, values]) => values.map((value) => [value, key]))
);

export const normalizePaymentStatus = (status = "") => {
    const normalized = String(status || "").toLowerCase();

    if (!normalized) {
        return "unknown";
    }

    if (mapStatus.has(normalized)) {
        return mapStatus.get(normalized);
    }

    return normalized;
};

export const getPaymentStatusLabel = (status) => {
    const normalized = normalizePaymentStatus(status);
    return STATUS_LABELS[normalized] || STATUS_LABELS.unknown;
};

export const getPaymentStatusIcon = (status) => {
    const normalized = normalizePaymentStatus(status);
    return STATUS_ICONS[normalized] || STATUS_ICONS.unknown;
};

export const getPaymentStatusBadgeClass = (status) => {
    const normalized = normalizePaymentStatus(status);
    return STATUS_BADGE_CLASSES[normalized] || STATUS_BADGE_CLASSES.unknown;
};
