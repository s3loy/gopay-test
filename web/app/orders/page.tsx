'use client';

import { useState } from 'react';
import { getOrder } from '@/lib/api';

export default function OrdersPage() {
  const [orderNo, setOrderNo] = useState('');
  const [order, setOrder] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleQuery = async () => {
    if (!orderNo.trim()) return;
    setLoading(true);
    setError('');
    try {
      const data = await getOrder(orderNo.trim());
      setOrder(data);
    } catch (err: any) {
      setError(err.message || '查询失败');
      setOrder(null);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-lg mx-auto">
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold mb-4">查询订单</h2>

        <div className="flex gap-2">
          <input
            type="text"
            placeholder="输入订单号"
            value={orderNo}
            onChange={(e) => setOrderNo(e.target.value)}
            className="flex-1 px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button
            onClick={handleQuery}
            disabled={loading}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            {loading ? '查询中...' : '查询'}
          </button>
        </div>

        {error && (
          <div className="mt-4 p-3 bg-red-50 text-red-700 rounded-md text-sm">{error}</div>
        )}

        {order && (
          <div className="mt-4 p-3 bg-gray-50 rounded-md">
            <pre className="text-xs text-gray-600 overflow-auto">{JSON.stringify(order, null, 2)}</pre>
          </div>
        )}
      </div>
    </div>
  );
}
