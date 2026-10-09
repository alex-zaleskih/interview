export class ApiError extends Error {
  constructor(status, message) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

/**
 * Calls the backend JSON API.
 * Redirects to the login page when the session is missing or expired.
 *
 * @param {string} path - e.g. '/api/orders?page=1'
 * @param {{method?: string, body?: unknown}} [options]
 * @returns {Promise<any>} parsed JSON response, or null for empty responses
 */
export async function api(path, { method = 'GET', body } = {}) {
  const response = await fetch(path, {
    method,
    headers: body !== undefined ? { 'Content-Type': 'application/json' } : {},
    body: body !== undefined ? JSON.stringify(body) : undefined,
    credentials: 'same-origin',
  });

  if (response.status === 401 && window.location.pathname !== '/login') {
    window.location.replace('/login');
    throw new ApiError(401, 'Требуется вход');
  }

  const data = response.status === 204 ? null : await response.json().catch(() => null);

  if (!response.ok) {
    throw new ApiError(response.status, data?.error ?? `Ошибка сервера (${response.status})`);
  }
  return data;
}
