'use client';

import { useState } from 'react';
import { createOrder, createPayment } from '@/lib/api';

export default function HomePage() {
  const [subject, setSubject] = useState('测试商品');
  const [amount, setAmount] = useState('100');
  const [channel, setChannel] = useState('wechat');
  const [method, setMethod] = useState('native');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [error, setError] = useState('');

  const handlePay = async () => {
    setLoading(true);
    setError('');
    setResult(null);

    try {
      const order = await createOrder({
        user_id: 1,
        subject,
        amount: parseInt(amount, 10),
        currency: 'CNY',
        expire_minutes: 30,
      });

      const payment = await createPayment({
        order_no: order.order_no,
        channel,
        method,
        client_ip: '127.0.0.1',
      });

      setResult(payment);
    } catch (err: any) {
      setError(err.message || '支付创建失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-lg mx-auto">
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold mb-4">创建订单</h2>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">商品名称</label>
            <input
              type="text"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">金额（分）</label>
            <input
              type="number"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">支付渠道</label>
            <select
              value={channel}
              onChange={(e) => {
                setChannel(e.target.value);
                setMethod(e.target.value === 'wechat' ? 'native' : 'pc');
              }}
              className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="wechat">微信支付</option>
              <option value="alipay">支付宝</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">支付方式</label>
            <select
              value={method}
              onChange={(e) => setMethod(e.target.value)}
              className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              {channel === 'wechat' ? (
                <>
                  <option value="native">扫码支付</option>
                  <option value="jsapi">JSAPI</option>
                </>
              ) : (
                <>
                  <option value="pc">PC 支付</option>
                  <option value="wap">WAP 支付</option>
                </>
              )}
            </select>
          </div>

          <button
            onClick={handlePay}
            disabled={loading}
            className="w-full bg-blue-600 text-white py-2 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? '处理中...' : '创建支付'}
          </button>
        </div>

        {error && (
          <div className="mt-4 p-3 bg-red-50 text-red-700 rounded-md text-sm">{error}</div>
        )}

        {result && (
          <div className="mt-4 p-3 bg-green-50 rounded-md">
            <p className="text-sm text-green-800 font-medium">支付创建成功</p>
            <pre className="mt-2 text-xs text-gray-600 overflow-auto">{JSON.stringify(result, null, 2)}</pre>
          </div>
        )}
      </div>
    </div>
  );
}
