// Seeds the "shop" database with demo users and orders.
//
// Runs automatically on the first start of the mongo container (empty volume).
// To re-run manually: make seed
//
// The data is generated with a fixed PRNG seed, so every run produces the same documents and IDs.

const shop = db.getSiblingDB('shop');
shop.users.drop();
shop.orders.drop();

// ---------- helpers ----------

function mulberry32(seed) {
  return function () {
    seed |= 0;
    seed = (seed + 0x6d2b79f5) | 0;
    let t = Math.imul(seed ^ (seed >>> 15), 1 | seed);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

const rand = mulberry32(20261001);
const randInt = (min, max) => min + Math.floor(rand() * (max - min + 1));
const pick = (arr) => arr[Math.floor(rand() * arr.length)];
const chance = (p) => rand() < p;

const MINUTE = 60 * 1000;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

// Deterministic ObjectId: 4-byte timestamp + 8-byte counter.
let idCounter = 1;
function objectIdAt(date) {
  const ts = Math.floor(date.getTime() / 1000).toString(16).padStart(8, '0');
  const counter = (idCounter++).toString(16).padStart(16, '0');
  return ObjectId(ts + counter);
}

// ---------- reference data ----------

const products = [
  { product_id: 'KB-1001', name: 'Механическая клавиатура TKL, коричневые свитчи', price: 749000 },
  { product_id: 'MS-2040', name: 'Беспроводная мышь эргономичная', price: 459000 },
  { product_id: 'MN-2701', name: 'Монитор 27" IPS 2560×1440', price: 2899000 },
  { product_id: 'HB-0710', name: 'USB-C хаб 7-в-1', price: 329000 },
  { product_id: 'HP-3300', name: 'Наушники с активным шумоподавлением', price: 1499000 },
  { product_id: 'WC-1080', name: 'Веб-камера Full HD с микрофоном', price: 389000 },
  { product_id: 'ST-0150', name: 'Подставка для ноутбука алюминиевая', price: 249000 },
  { product_id: 'CB-0200', name: 'Кабель USB-C — USB-C 2 м, 100 Вт', price: 79000 },
  { product_id: 'SD-1000', name: 'SSD NVMe 1 ТБ', price: 899000 },
  { product_id: 'MP-0900', name: 'Коврик для мыши XL 900×400', price: 149000 },
  { product_id: 'LM-0450', name: 'Настольная LED-лампа с регулировкой яркости', price: 349000 },
  { product_id: 'BP-1560', name: 'Рюкзак для ноутбука 15.6"', price: 549000 },
  { product_id: 'PB-2000', name: 'Внешний аккумулятор 20000 мА·ч', price: 299000 },
  { product_id: 'SP-0500', name: 'Портативная Bluetooth-колонка', price: 499000 },
  { product_id: 'CH-0065', name: 'Зарядное устройство GaN 65 Вт', price: 279000 },
];

const pickupPoints = [
  { code: 'MSK-014', name: 'Пункт выдачи «Тверская»', address: 'Москва, ул. Тверская, д. 12', working_hours: 'Ежедневно 10:00–21:00' },
  { code: 'MSK-027', name: 'Пункт выдачи «Курская»', address: 'Москва, ул. Земляной Вал, д. 33', working_hours: 'Пн–Сб 09:00–20:00' },
  { code: 'SPB-003', name: 'Пункт выдачи «Невский»', address: 'Санкт-Петербург, Невский пр-т, д. 88', working_hours: 'Ежедневно 10:00–22:00' },
];

const customerComments = [
  'Позвонить за час до доставки',
  'Домофон не работает, звоните на телефон',
  'Оставить у двери',
  'Нужен чек для юрлица',
];

const cancelReasons = [
  'Отменён покупателем',
  'Товара нет в наличии',
  'Не удалось связаться с покупателем',
];

const COURIER_PRICE = 39000;
const POST_PRICE = 29000;
const FREE_COURIER_FROM = 500000;

// ---------- users ----------

const users = [
  {
    _id: ObjectId('652f1a000000000000000001'),
    email: 'ivan@example.com',
    full_name: 'Иван Петров',
    phone: '+7 (900) 123-45-67',
    created_at: new Date('2024-03-12T09:15:00Z'),
    city: 'Москва',
    address: { city: 'Москва', street: 'ул. Профсоюзная', house: '45', apartment: '112', postal_code: '117420' },
    ordersCount: 23,
  },
  {
    _id: ObjectId('652f1a000000000000000002'),
    email: 'maria@example.com',
    full_name: 'Мария Смирнова',
    phone: '+7 (921) 555-12-34',
    created_at: new Date('2025-01-20T18:40:00Z'),
    city: 'Санкт-Петербург',
    address: { city: 'Санкт-Петербург', street: 'ул. Марата', house: '9', apartment: '4', postal_code: '191025' },
    ordersCount: 6,
  },
  {
    _id: ObjectId('652f1a000000000000000003'),
    email: 'alexey@example.com',
    full_name: 'Алексей Новиков',
    phone: '+7 (916) 777-00-11',
    created_at: new Date('2026-09-30T08:00:00Z'),
    city: 'Москва',
    address: null,
    ordersCount: 0,
  },
];

shop.users.insertMany(
  users.map((u) => ({
    _id: u._id,
    email: u.email,
    full_name: u.full_name,
    phone: u.phone,
    created_at: u.created_at,
  })),
);

// ---------- orders ----------

function buildItems() {
  const count = randInt(1, 4);
  const pool = products.slice();
  const items = [];
  for (let i = 0; i < count; i++) {
    const product = pool.splice(Math.floor(rand() * pool.length), 1)[0];
    items.push({
      product_id: product.product_id,
      name: product.name,
      quantity: chance(0.2) ? randInt(2, 3) : 1,
      unit_price: product.price,
    });
  }
  return items;
}

function buildDelivery(user, subtotal) {
  const method = pick(['courier', 'courier', 'pickup', 'post']);
  switch (method) {
    case 'courier':
      return {
        method,
        price: subtotal >= FREE_COURIER_FROM ? 0 : COURIER_PRICE,
        address: user.address,
      };
    case 'pickup': {
      const cityPoints = pickupPoints.filter((p) => p.address.startsWith(user.city));
      return { method, price: 0, pickup_point: pick(cityPoints) };
    }
    default:
      return { method, price: POST_PRICE, address: user.address };
  }
}

function buildOrder(user, createdAt, status) {
  const items = buildItems();
  const subtotal = items.reduce((sum, it) => sum + it.quantity * it.unit_price, 0);
  const delivery = buildDelivery(user, subtotal);

  let paymentMethod = pick(['card', 'card', 'sbp', 'cash']);
  if (paymentMethod === 'cash' && (status === 'paid' || delivery.method === 'post')) {
    paymentMethod = 'card'; // cash is accepted on delivery only, and not by post
  }

  const history = [{ status: 'new', at: createdAt }];
  let at = createdAt.getTime();
  let paidAt = null;

  if (status === 'cancelled') {
    at += randInt(1, 20) * HOUR;
    history.push({ status: 'cancelled', at: new Date(at), comment: pick(cancelReasons) });
  } else {
    if (status !== 'new' && paymentMethod !== 'cash') {
      at += randInt(1, 15) * MINUTE;
      paidAt = new Date(at);
      history.push({ status: 'paid', at: paidAt });
    }
    if (status === 'shipped' || status === 'delivered') {
      at += randInt(18, 40) * HOUR;
      history.push({ status: 'shipped', at: new Date(at) });
      if (delivery.method === 'post') {
        delivery.tracking_number = 'RA' + String(randInt(100000000, 999999999)) + 'RU';
      }
    }
    if (status === 'delivered') {
      at += randInt(1, 4) * DAY;
      history.push({ status: 'delivered', at: new Date(at) });
      if (paymentMethod === 'cash') paidAt = new Date(at);
    }
  }

  const order = {
    _id: objectIdAt(createdAt),
    user_id: user._id,
    status,
    created_at: createdAt,
    updated_at: history[history.length - 1].at,
    currency: 'RUB',
    items: items.map((it) => ({
      product_id: it.product_id,
      name: it.name,
      quantity: NumberInt(it.quantity),
      unit_price: NumberLong(it.unit_price),
    })),
    subtotal: NumberLong(subtotal),
    delivery: { ...delivery, price: NumberLong(delivery.price) },
    payment: { method: paymentMethod, paid_at: paidAt },
    status_history: history,
  };

  let discount = 0;
  if (chance(0.25)) {
    discount = Math.round((subtotal * 0.1) / 100) * 100;
    order.promo_code = 'AUTUMN10';
    order.discount = NumberLong(discount);
  }
  order.total = NumberLong(subtotal - discount + delivery.price);

  if (chance(0.3)) {
    order.comment = pick(customerComments);
  }
  return order;
}

function statusForPosition(i) {
  // i = 0 is the newest order: the freshest orders are still in progress, older ones are finished.
  if (i === 0) return 'new';
  if (i === 1) return 'paid';
  if (i === 2) return 'shipped';
  return chance(0.15) ? 'cancelled' : 'delivered';
}

const orders = [];
const newestOrderAt = new Date('2026-10-08T11:20:00Z').getTime();

for (const user of users) {
  let createdAt = newestOrderAt - randInt(0, 2) * DAY - randInt(0, 600) * MINUTE;
  for (let i = 0; i < user.ordersCount; i++) {
    orders.push(buildOrder(user, new Date(createdAt), statusForPosition(i)));
    createdAt -= randInt(4, 21) * DAY + randInt(0, 720) * MINUTE;
  }
}

// Human-readable order numbers are sequential in time across all users.
orders.sort((a, b) => a.created_at - b.created_at);
orders.forEach((o, i) => {
  o.number = 'ORD-' + o.created_at.getUTCFullYear() + '-' + String(1000 + i * 7 + 3).padStart(5, '0');
});

shop.orders.insertMany(orders);

print(`Seeded ${shop.users.countDocuments()} users and ${shop.orders.countDocuments()} orders`);
