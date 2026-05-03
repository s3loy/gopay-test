export interface Order {
  order_no: string;
  user_id: number;
  subject: string;
  amount: number;
  currency: string;
  status: string;
  expired_at: number;
  paid_at?: number;
  description?: string;
  created_at: number;
}

export interface Payment {
  payment_no: string;
  order_no: string;
  channel: string;
  method: string;
  amount: number;
  currency: string;
  status: string;
  pay_params?: Record<string, unknown>;
  expire_at: number;
  created_at: number;
}

export interface Refund {
  refund_no: string;
  payment_no: string;
  order_no: string;
  channel: string;
  amount: number;
  reason: string;
  status: string;
  created_at: number;
}
