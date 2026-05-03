const API_BASE = '';

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: {
      'Content-Type': 'application/json',
    },
    ...options,
  });

  const data = await res.json();
  if (data.code !== 0) {
    throw new Error(data.message || '请求失败');
  }
  return data.data;
}

export function createOrder(body: {
  user_id: number;
  subject: string;
  amount: number;
  currency?: string;
  description?: string;
  expire_minutes?: number;
}) {
  return request('/api/v1/orders', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

export function createPayment(body: {
  order_no: string;
  channel: string;
  method: string;
  client_ip: string;
  notify_url?: string;
  return_url?: string;
  openid?: string;
  buyer_id?: string;
  expire_minutes?: number;
}) {
  return request('/api/v1/payments', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

export function getOrder(orderNo: string) {
  return request(`/api/v1/orders/${orderNo}`);
}

export function getPayment(paymentNo: string) {
  return request(`/api/v1/payments/${paymentNo}`);
}
