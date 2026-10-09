import { el } from './dom.js';

const ORDER_STATUSES = {
  new: { label: 'Новый', tone: 'info' },
  paid: { label: 'Оплачен', tone: 'accent' },
  shipped: { label: 'В пути', tone: 'warning' },
  delivered: { label: 'Доставлен', tone: 'success' },
  cancelled: { label: 'Отменён', tone: 'muted' },
};

export function statusLabel(status) {
  return ORDER_STATUSES[status]?.label ?? status;
}

export function statusBadge(status) {
  const tone = ORDER_STATUSES[status]?.tone ?? 'muted';
  return el('span', { className: `badge badge--${tone}` }, statusLabel(status));
}

/** Formats money from the API: amount is in minor units (kopecks). */
export function formatMoney({ amount, currency }) {
  return new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency,
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(amount / 100);
}

const dateFormat = new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' });
const dateTimeFormat = new Intl.DateTimeFormat('ru-RU', {
  day: 'numeric',
  month: 'long',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
});

export function formatDate(iso) {
  return dateFormat.format(new Date(iso));
}

export function formatDateTime(iso) {
  return dateTimeFormat.format(new Date(iso));
}

/** pluralize(5, ['заказ', 'заказа', 'заказов']) → 'заказов' */
export function pluralize(n, [one, few, many]) {
  const mod10 = n % 10;
  const mod100 = n % 100;
  if (mod10 === 1 && mod100 !== 11) return one;
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return few;
  return many;
}
