'use client';

import React, { useState, useEffect } from 'react';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

interface InventoryItem {
  id: string;
  sku: string;
  name: string;
  description: string;
  quantity: number;
  price_cents: number;
  created_at: string;
}

interface Transaction {
  id: string;
  customer_id: string;
  amount_cents: number;
  currency: string;
  status: string;
  provider: string;
  provider_tx_id: string;
  created_at: string;
}

interface JournalEntry {
  id: string;
  reference_id: string;
  memo: string;
  amount_cents: number;
  currency: string;
  status: string;
  provider: string;
  sync_log: string;
  created_at: string;
}

export default function Dashboard() {
  const [activeTab, setActiveTab] = useState<'inventory' | 'payment' | 'accounting'>('inventory');
  
  // Data states
  const [items, setItems] = useState<InventoryItem[]>([]);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [entries, setEntries] = useState<JournalEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ text: string; type: 'success' | 'error' } | null>(null);

  // Forms
  const [newItem, setNewItem] = useState({ sku: '', name: '', description: '', quantity: 10, price: '19.99' });
  const [newPay, setNewPay] = useState({ customerId: 'cust_user_1', amount: '49.99', currency: 'USD', provider: 'stripe' });

  const fetchAll = async () => {
    setLoading(true);
    try {
      const [resInv, resPay, resAcc] = await Promise.all([
        fetch(`${API_BASE}/inventory/items`).then(r => r.json()).catch(() => ({ items: [] })),
        fetch(`${API_BASE}/payment/transactions`).then(r => r.json()).catch(() => ({ transactions: [] })),
        fetch(`${API_BASE}/accounting/entries`).then(r => r.json()).catch(() => ({ entries: [] }))
      ]);

      setItems(resInv.items || []);
      setTransactions(resPay.transactions || []);
      setEntries(resAcc.entries || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAll();
  }, []);

  const showMsg = (text: string, type: 'success' | 'error' = 'success') => {
    setMessage({ text, type });
    setTimeout(() => setMessage(null), 4000);
  };

  // Inventory Handlers
  const handleCreateItem = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await fetch(`${API_BASE}/inventory/items`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          sku: newItem.sku,
          name: newItem.name,
          description: newItem.description,
          quantity: Number(newItem.quantity),
          price_cents: Math.round(parseFloat(newItem.price) * 100)
        })
      });
      if (!res.ok) throw new Error(await res.text());
      showMsg('Inventory item created successfully!');
      setNewItem({ sku: '', name: '', description: '', quantity: 10, price: '19.99' });
      fetchAll();
    } catch (err: any) {
      showMsg(err.message || 'Failed to create item', 'error');
    }
  };

  const handleAdjustStock = async (id: string, delta: number) => {
    try {
      const res = await fetch(`${API_BASE}/inventory/items/${id}/adjust`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ delta })
      });
      if (!res.ok) throw new Error(await res.text());
      showMsg(`Stock updated (${delta > 0 ? '+' : ''}${delta})`);
      fetchAll();
    } catch (err: any) {
      showMsg(err.message || 'Failed to adjust stock', 'error');
    }
  };

  // Payment Handler
  const handleProcessPayment = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await fetch(`${API_BASE}/payment/charge`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          customer_id: newPay.customerId,
          amount_cents: Math.round(parseFloat(newPay.amount) * 100),
          currency: newPay.currency,
          provider: newPay.provider
        })
      });
      if (!res.ok) throw new Error(await res.text());
      showMsg(`Payment charged using ${newPay.provider.toUpperCase()} strategy! Accounting auto-synced via Domain Events.`);
      fetchAll();
    } catch (err: any) {
      showMsg(err.message || 'Failed to process payment', 'error');
    }
  };

  return (
    <div style={{ maxWidth: '1200px', margin: '0 auto', padding: '2rem 1rem' }}>
      {/* Header */}
      <header style={{ marginBottom: '2.5rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1 style={{ fontSize: '2rem', fontWeight: 800, margin: 0, background: 'linear-gradient(to right, #818cf8, #c084fc)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent' }}>
            Modular Monolith Dashboard
          </h1>
          <p style={{ color: '#9ca3af', margin: '0.4rem 0 0 0' }}>
            Go (Gin + GORM) Backend • SQLite/Postgres Single DB • Next.js Frontend
          </p>
        </div>
        <button onClick={fetchAll} className="btn-secondary" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          🔄 Refresh Data
        </button>
      </header>

      {/* Message Banner */}
      {message && (
        <div style={{
          padding: '1rem',
          borderRadius: '8px',
          marginBottom: '1.5rem',
          background: message.type === 'success' ? 'rgba(16, 185, 129, 0.2)' : 'rgba(239, 68, 68, 0.2)',
          border: `1px solid ${message.type === 'success' ? '#10b981' : '#ef4444'}`,
          color: message.type === 'success' ? '#6ee7b7' : '#fca5a5'
        }}>
          {message.text}
        </div>
      )}

      {/* Module Tabs */}
      <div style={{ display: 'flex', gap: '1rem', marginBottom: '2rem', borderBottom: '1px solid rgba(255,255,255,0.1)', paddingBottom: '0.5rem' }}>
        <button
          onClick={() => setActiveTab('inventory')}
          style={{
            background: 'none',
            border: 'none',
            color: activeTab === 'inventory' ? '#818cf8' : '#9ca3af',
            fontSize: '1.1rem',
            fontWeight: 600,
            cursor: 'pointer',
            padding: '0.5rem 1rem',
            borderBottom: activeTab === 'inventory' ? '2px solid #818cf8' : '2px solid transparent'
          }}
        >
          📦 Inventory Management ({items.length})
        </button>
        <button
          onClick={() => setActiveTab('payment')}
          style={{
            background: 'none',
            border: 'none',
            color: activeTab === 'payment' ? '#818cf8' : '#9ca3af',
            fontSize: '1.1rem',
            fontWeight: 600,
            cursor: 'pointer',
            padding: '0.5rem 1rem',
            borderBottom: activeTab === 'payment' ? '2px solid #818cf8' : '2px solid transparent'
          }}
        >
          💳 Payments Strategy ({transactions.length})
        </button>
        <button
          onClick={() => setActiveTab('accounting')}
          style={{
            background: 'none',
            border: 'none',
            color: activeTab === 'accounting' ? '#818cf8' : '#9ca3af',
            fontSize: '1.1rem',
            fontWeight: 600,
            cursor: 'pointer',
            padding: '0.5rem 1rem',
            borderBottom: activeTab === 'accounting' ? '2px solid #818cf8' : '2px solid transparent'
          }}
        >
          📊 Accounting Ledger ({entries.length})
        </button>
      </div>

      {/* Tab 1: Inventory */}
      {activeTab === 'inventory' && (
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 2fr', gap: '2rem' }}>
          {/* Create Form */}
          <div className="glass-card" style={{ padding: '1.5rem' }}>
            <h3 style={{ margin: '0 0 1rem 0', color: '#f3f4f6' }}>Create New Item</h3>
            <form onSubmit={handleCreateItem} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.85rem', color: '#9ca3af', marginBottom: '0.3rem' }}>SKU Code</label>
                <input
                  type="text"
                  required
                  placeholder="SKU-1001"
                  value={newItem.sku}
                  onChange={e => setNewItem({ ...newItem, sku: e.target.value })}
                  style={{ width: '100%', padding: '0.5rem', borderRadius: '6px', background: 'rgba(0,0,0,0.3)', border: '1px solid rgba(255,255,255,0.1)', color: '#fff' }}
                />
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.85rem', color: '#9ca3af', marginBottom: '0.3rem' }}>Item Name</label>
                <input
                  type="text"
                  required
                  placeholder="Wireless Mouse"
                  value={newItem.name}
                  onChange={e => setNewItem({ ...newItem, name: e.target.value })}
                  style={{ width: '100%', padding: '0.5rem', borderRadius: '6px', background: 'rgba(0,0,0,0.3)', border: '1px solid rgba(255,255,255,0.1)', color: '#fff' }}
                />
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.85rem', color: '#9ca3af', marginBottom: '0.3rem' }}>Description</label>
                <input
                  type="text"
                  placeholder="Ergonomic optical mouse"
                  value={newItem.description}
                  onChange={e => setNewItem({ ...newItem, description: e.target.value })}
                  style={{ width: '100%', padding: '0.5rem', borderRadius: '6px', background: 'rgba(0,0,0,0.3)', border: '1px solid rgba(255,255,255,0.1)', color: '#fff' }}
                />
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.85rem', color: '#9ca3af', marginBottom: '0.3rem' }}>Quantity</label>
                  <input
                    type="number"
                    required
                    value={newItem.quantity}
                    onChange={e => setNewItem({ ...newItem, quantity: Number(e.target.value) })}
                    style={{ width: '100%', padding: '0.5rem', borderRadius: '6px', background: 'rgba(0,0,0,0.3)', border: '1px solid rgba(255,255,255,0.1)', color: '#fff' }}
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.85rem', color: '#9ca3af', marginBottom: '0.3rem' }}>Price ($)</label>
                  <input
                    type="number"
                    step="0.01"
                    required
                    value={newItem.price}
                    onChange={e => setNewItem({ ...newItem, price: e.target.value })}
                    style={{ width: '100%', padding: '0.5rem', borderRadius: '6px', background: 'rgba(0,0,0,0.3)', border: '1px solid rgba(255,255,255,0.1)', color: '#fff' }}
                  />
                </div>
              </div>
              <button type="submit" className="btn-primary" style={{ marginTop: '0.5rem' }}>Save Item</button>
            </form>
          </div>

          {/* List View */}
          <div className="glass-card" style={{ padding: '1.5rem' }}>
            <h3 style={{ margin: '0 0 1rem 0', color: '#f3f4f6' }}>Inventory Items Table (`inventory_items`)</h3>
            {items.length === 0 ? (
              <p style={{ color: '#6b7280' }}>No inventory items found. Add one above!</p>
            ) : (
              <div style={{ overflowX: 'auto' }}>
                <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left' }}>
                  <thead>
                    <tr style={{ borderBottom: '1px solid rgba(255,255,255,0.1)', color: '#9ca3af', fontSize: '0.85rem' }}>
                      <th style={{ padding: '0.5rem' }}>SKU</th>
                      <th style={{ padding: '0.5rem' }}>Name</th>
                      <th style={{ padding: '0.5rem' }}>Price</th>
                      <th style={{ padding: '0.5rem' }}>Stock</th>
                      <th style={{ padding: '0.5rem' }}>Adjust</th>
                    </tr>
                  </thead>
                  <tbody>
                    {items.map(item => (
                      <tr key={item.id} style={{ borderBottom: '1px solid rgba(255,255,255,0.05)' }}>
                        <td style={{ padding: '0.75rem 0.5rem', fontFamily: 'monospace', color: '#818cf8' }}>{item.sku}</td>
                        <td style={{ padding: '0.75rem 0.5rem' }}>
                          <div style={{ fontWeight: 600 }}>{item.name}</div>
                          <div style={{ fontSize: '0.8rem', color: '#9ca3af' }}>{item.description}</div>
                        </td>
                        <td style={{ padding: '0.75rem 0.5rem' }}>${(item.price_cents / 100).toFixed(2)}</td>
                        <td style={{ padding: '0.75rem 0.5rem' }}>
                          <span className={item.quantity > 5 ? 'badge-emerald' : 'badge-purple'}>
                            {item.quantity} in stock
                          </span>
                        </td>
                        <td style={{ padding: '0.75rem 0.5rem' }}>
                          <div style={{ display: 'flex', gap: '0.3rem' }}>
                            <button onClick={() => handleAdjustStock(item.id, 1)} className="btn-secondary" style={{ padding: '0.2rem 0.6rem' }}>+</button>
                            <button onClick={() => handleAdjustStock(item.id, -1)} className="btn-secondary" style={{ padding: '0.2rem 0.6rem' }}>-</button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Tab 2: Payments */}
      {activeTab === 'payment' && (
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 2fr', gap: '2rem' }}>
          {/* Charge Form */}
          <div className="glass-card" style={{ padding: '1.5rem' }}>
            <h3 style={{ margin: '0 0 1rem 0', color: '#f3f4f6' }}>Process Charge via Strategy</h3>
            <form onSubmit={handleProcessPayment} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.85rem', color: '#9ca3af', marginBottom: '0.3rem' }}>Customer ID</label>
                <input
                  type="text"
                  required
                  value={newPay.customerId}
                  onChange={e => setNewPay({ ...newPay, customerId: e.target.value })}
                  style={{ width: '100%', padding: '0.5rem', borderRadius: '6px', background: 'rgba(0,0,0,0.3)', border: '1px solid rgba(255,255,255,0.1)', color: '#fff' }}
                />
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.85rem', color: '#9ca3af', marginBottom: '0.3rem' }}>Amount ($)</label>
                  <input
                    type="number"
                    step="0.01"
                    required
                    value={newPay.amount}
                    onChange={e => setNewPay({ ...newPay, amount: e.target.value })}
                    style={{ width: '100%', padding: '0.5rem', borderRadius: '6px', background: 'rgba(0,0,0,0.3)', border: '1px solid rgba(255,255,255,0.1)', color: '#fff' }}
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.85rem', color: '#9ca3af', marginBottom: '0.3rem' }}>Currency</label>
                  <input
                    type="text"
                    required
                    value={newPay.currency}
                    onChange={e => setNewPay({ ...newPay, currency: e.target.value })}
                    style={{ width: '100%', padding: '0.5rem', borderRadius: '6px', background: 'rgba(0,0,0,0.3)', border: '1px solid rgba(255,255,255,0.1)', color: '#fff' }}
                  />
                </div>
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.85rem', color: '#9ca3af', marginBottom: '0.3rem' }}>Payment Strategy</label>
                <select
                  value={newPay.provider}
                  onChange={e => setNewPay({ ...newPay, provider: e.target.value })}
                  style={{ width: '100%', padding: '0.5rem', borderRadius: '6px', background: 'rgba(0,0,0,0.3)', border: '1px solid rgba(255,255,255,0.1)', color: '#fff' }}
                >
                  <option value="stripe" style={{ background: '#1e1b4b' }}>StripeStrategy (Default)</option>
                </select>
              </div>
              <button type="submit" className="btn-primary" style={{ marginTop: '0.5rem' }}>Process Payment</button>
            </form>
          </div>

          {/* Transactions Table */}
          <div className="glass-card" style={{ padding: '1.5rem' }}>
            <h3 style={{ margin: '0 0 1rem 0', color: '#f3f4f6' }}>Payment Transactions (`pay_transactions`)</h3>
            {transactions.length === 0 ? (
              <p style={{ color: '#6b7280' }}>No transactions recorded yet.</p>
            ) : (
              <div style={{ overflowX: 'auto' }}>
                <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left' }}>
                  <thead>
                    <tr style={{ borderBottom: '1px solid rgba(255,255,255,0.1)', color: '#9ca3af', fontSize: '0.85rem' }}>
                      <th style={{ padding: '0.5rem' }}>Tx ID</th>
                      <th style={{ padding: '0.5rem' }}>Customer</th>
                      <th style={{ padding: '0.5rem' }}>Amount</th>
                      <th style={{ padding: '0.5rem' }}>Provider</th>
                      <th style={{ padding: '0.5rem' }}>Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {transactions.map(tx => (
                      <tr key={tx.id} style={{ borderBottom: '1px solid rgba(255,255,255,0.05)' }}>
                        <td style={{ padding: '0.75rem 0.5rem', fontFamily: 'monospace', fontSize: '0.8rem', color: '#9ca3af' }}>{tx.id.slice(0, 8)}...</td>
                        <td style={{ padding: '0.75rem 0.5rem', fontWeight: 600 }}>{tx.customer_id}</td>
                        <td style={{ padding: '0.75rem 0.5rem' }}>${(tx.amount_cents / 100).toFixed(2)} {tx.currency}</td>
                        <td style={{ padding: '0.75rem 0.5rem' }}>
                          <span className="badge-purple">{tx.provider}</span>
                        </td>
                        <td style={{ padding: '0.75rem 0.5rem' }}>
                          <span className="badge-emerald">{tx.status}</span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Tab 3: Accounting */}
      {activeTab === 'accounting' && (
        <div className="glass-card" style={{ padding: '1.5rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
            <h3 style={{ margin: 0, color: '#f3f4f6' }}>General Ledger Entries (`acc_entries`)</h3>
            <span className="badge-purple">Strategy: RilletStrategy</span>
          </div>
          <p style={{ color: '#9ca3af', fontSize: '0.9rem', marginBottom: '1.5rem' }}>
            Entries in this table are automatically generated via <strong>Domain Events (`payment.completed`)</strong> when a payment is processed.
          </p>

          {entries.length === 0 ? (
            <p style={{ color: '#6b7280' }}>No accounting entries recorded. Try processing a payment!</p>
          ) : (
            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid rgba(255,255,255,0.1)', color: '#9ca3af', fontSize: '0.85rem' }}>
                    <th style={{ padding: '0.5rem' }}>Entry ID</th>
                    <th style={{ padding: '0.5rem' }}>Ref Tx ID</th>
                    <th style={{ padding: '0.5rem' }}>Memo</th>
                    <th style={{ padding: '0.5rem' }}>Amount</th>
                    <th style={{ padding: '0.5rem' }}>Sync Status</th>
                    <th style={{ padding: '0.5rem' }}>Sync Log</th>
                  </tr>
                </thead>
                <tbody>
                  {entries.map(entry => (
                    <tr key={entry.id} style={{ borderBottom: '1px solid rgba(255,255,255,0.05)' }}>
                      <td style={{ padding: '0.75rem 0.5rem', fontFamily: 'monospace', fontSize: '0.8rem', color: '#9ca3af' }}>{entry.id.slice(0, 8)}...</td>
                      <td style={{ padding: '0.75rem 0.5rem', fontFamily: 'monospace', fontSize: '0.8rem', color: '#818cf8' }}>{entry.reference_id.slice(0, 8)}...</td>
                      <td style={{ padding: '0.75rem 0.5rem' }}>{entry.memo}</td>
                      <td style={{ padding: '0.75rem 0.5rem' }}>${(entry.amount_cents / 100).toFixed(2)} {entry.currency}</td>
                      <td style={{ padding: '0.75rem 0.5rem' }}>
                        <span className="badge-emerald">{entry.status}</span>
                      </td>
                      <td style={{ padding: '0.75rem 0.5rem', fontSize: '0.8rem', color: '#9ca3af' }}>{entry.sync_log}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
