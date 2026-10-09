import { api } from './api.js';

const form = document.getElementById('login-form');
const emailInput = document.getElementById('email');
const submitButton = document.getElementById('login-submit');
const errorBox = document.getElementById('login-error');

function showError(message) {
  errorBox.textContent = message;
  errorBox.hidden = !message;
}

form.addEventListener('submit', async (event) => {
  event.preventDefault();
  showError('');

  const email = emailInput.value.trim();
  if (!email) {
    showError('Введите email');
    return;
  }

  submitButton.disabled = true;
  try {
    await api('/api/auth/login', { method: 'POST', body: { email } });
    window.location.replace('/');
  } catch (err) {
    showError(err.status === 401 ? 'Пользователь с таким email не найден' : err.message);
  } finally {
    submitButton.disabled = false;
  }
});

document.querySelectorAll('[data-email]').forEach((chip) => {
  chip.addEventListener('click', () => {
    emailInput.value = chip.dataset.email;
    emailInput.focus();
  });
});

// Already logged in? Go straight to the cabinet.
api('/api/me')
  .then(() => window.location.replace('/'))
  .catch(() => {});
