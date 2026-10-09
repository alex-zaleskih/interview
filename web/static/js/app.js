import { api } from './api.js';
import { clear, el } from './dom.js';
import { formatDate, formatDateTime, formatMoney, pluralize, statusBadge } from './format.js';

const PAGE_LIMIT = 10;

const state = {
  page: 1,
  totalPages: 1,
};

const $ = (id) => document.getElementById(id);

// ---------- profile ----------

async function loadProfile() {
  try {
    const user = await api('/api/me');
    $('topbar-user-name').textContent = user.full_name;
    renderProfile(user);
  } catch (err) {
    $('profile').replaceChildren(el('p', { className: 'form-error' }, err.message));
  }
}

function renderProfile(user) {
  const initials = user.full_name
    .split(' ')
    .map((part) => part[0])
    .join('')
    .slice(0, 2);

  $('profile').replaceChildren(
    el('div', { className: 'profile__avatar' }, initials),
    el('div', { className: 'profile__name' }, user.full_name),
    el('dl', { className: 'profile__details' }, [
      el('dt', {}, 'Email'),
      el('dd', {}, user.email),
      el('dt', {}, 'Телефон'),
      el('dd', {}, user.phone || '—'),
      el('dt', {}, 'Клиент с'),
      el('dd', {}, formatDate(user.created_at)),
    ]),
  );
}

// ---------- orders ----------

async function loadOrders(page) {
  showOrdersState('Загрузка заказов…');
  try {
    const data = await api(`/api/orders?page=${page}&limit=${PAGE_LIMIT}`);
    renderOrders(data);
  } catch (err) {
    showOrdersState(err.message, { isError: true, onRetry: () => loadOrders(page) });
  }
}

function renderOrders({ items, total, page, limit }) {
  state.page = page;
  state.totalPages = Math.max(1, Math.ceil(total / limit));

  $('orders-count').textContent = total > 0 ? `${total} ${pluralize(total, ['заказ', 'заказа', 'заказов'])}` : '';

  if (total === 0) {
    showOrdersState('У вас пока нет заказов');
    return;
  }

  const tbody = $('orders-body');
  clear(tbody);
  for (const order of items) {
    tbody.append(renderOrderRow(order));
  }

  hideOrdersState();
  $('orders-table-wrapper').hidden = false;
  renderPagination();
}

function renderOrderRow(order) {
  return el('tr', { dataset: { orderId: order.id } }, [
    el('td', { className: 'order-number' }, order.number),
    el('td', { title: formatDateTime(order.created_at) }, formatDate(order.created_at)),
    el('td', {}, statusBadge(order.status)),
    el('td', { className: 'num' }, order.items_count),
    el('td', { className: 'num' }, formatMoney(order.total)),
  ]);
}

function renderPagination() {
  const visible = state.totalPages > 1;
  $('pagination').hidden = !visible;
  if (!visible) return;

  $('page-info').textContent = `Страница ${state.page} из ${state.totalPages}`;
  $('prev-page').disabled = state.page <= 1;
  $('next-page').disabled = state.page >= state.totalPages;
}

function showOrdersState(message, { isError = false, onRetry } = {}) {
  $('orders-table-wrapper').hidden = true;
  $('pagination').hidden = true;

  const box = $('orders-state');
  box.hidden = false;
  box.classList.toggle('state--error', isError);
  box.replaceChildren(el('p', {}, message));
  if (onRetry) {
    box.append(el('button', { className: 'button button--ghost', type: 'button', onClick: onRetry }, 'Повторить'));
  }
}

function hideOrdersState() {
  $('orders-state').hidden = true;
}

// ---------- init ----------

async function logout() {
  await api('/api/auth/logout', { method: 'POST' });
  window.location.replace('/login');
}

$('logout-button').addEventListener('click', logout);
$('prev-page').addEventListener('click', () => loadOrders(state.page - 1));
$('next-page').addEventListener('click', () => loadOrders(state.page + 1));

loadProfile();
loadOrders(1);
