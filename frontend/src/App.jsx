import { useState, useEffect } from 'react'
import './App.css'

function App() {
  const [activeTab, setActiveTab] = useState('orders')

  return (
    <div className="container">
      <header className="header">
        <h1>🛒 E-Commerce Dashboard</h1>
        <div className="tabs">
          <button
            className={activeTab === 'orders' ? 'active' : ''}
            onClick={() => setActiveTab('orders')}
          >
            📦 Orders
          </button>
          <button
            className={activeTab === 'inventory' ? 'active' : ''}
            onClick={() => setActiveTab('inventory')}
          >
            🏭 Inventory
          </button>
          <button
            className={activeTab === 'payments' ? 'active' : ''}
            onClick={() => setActiveTab('payments')}
          >
            💳 Payments
          </button>
        </div>
      </header>

      <main className="content">
        {activeTab === 'orders' && <OrdersPanel />}
        {activeTab === 'inventory' && <InventoryPanel />}
        {activeTab === 'payments' && <PaymentsPanel />}
      </main>
    </div>
  )
}

// --- Order Component ---
function OrdersPanel() {
  const [orders, setOrders] = useState([])
  const [loading, setLoading] = useState(false)
  const [createForm, setCreateForm] = useState({
    user_id: 1,
    product_id: "101",
    quantity: 1,
    price: 100
  })

  const fetchOrders = async () => {
    setLoading(true)
    try {
      const res = await fetch('/api/v1/orders')
      if (!res.ok) {
        // Now backend returns 200 [] for empty list, so !res.ok is a real error
        const errData = await res.json().catch(() => ({}))
        throw new Error(errData.msg || `Server error: ${res.status}`)
      }
      const data = await res.json()
      setOrders(data.data || [])
    } catch (err) {
      console.error("Fetch error:", err)
      alert(`Failed to fetch orders: ${err.message}`) // Show error as requested
    } finally {
      setLoading(false)
    }
  }

  const createOrder = async () => {
    try {
      const res = await fetch('/api/v1/orders', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          user_id: Number(createForm.user_id),
          items: [{
            product_id: createForm.product_id,
            quantity: Number(createForm.quantity),
            price: Number(createForm.price)
          }]
        })
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.msg || 'Error creating order')
      alert(`Order Created! ID: ${data.data.order_id}`)
      fetchOrders()
    } catch (err) {
      alert(err.message)
    }
  }

  useEffect(() => { fetchOrders() }, [])

  return (
    <div className="panel">
      <div className="card">
        <h2>New Order</h2>
        <div className="form-grid">
          <label>
            User ID:
            <input type="number" value={createForm.user_id} onChange={e => setCreateForm({ ...createForm, user_id: e.target.value })} />
          </label>
          <label>
            Product ID:
            <input type="text" value={createForm.product_id} onChange={e => setCreateForm({ ...createForm, product_id: e.target.value })} />
          </label>
          <label>
            Quantity:
            <input type="number" value={createForm.quantity} onChange={e => setCreateForm({ ...createForm, quantity: e.target.value })} />
          </label>
          <label>
            Price (cents):
            <input type="number" value={createForm.price} onChange={e => setCreateForm({ ...createForm, price: e.target.value })} />
          </label>
          <button onClick={createOrder}>Create Order</button>
        </div>
      </div>

      <div className="card">
        <div className="flex-header">
          <h2>Order List</h2>
          <button onClick={fetchOrders} disabled={loading}>🔄 Refresh</button>
        </div>
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>User</th>
              <th>Amount</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {orders.map(o => (
              <tr key={o.id}>
                <td>{o.id}</td>
                <td>{o.user_id}</td>
                <td>${(o.total_amount / 100).toFixed(2)}</td>
                <td><span className={`status ${o.status}`}>{o.status}</span></td>
              </tr>
            ))}
            {orders.length === 0 && <tr><td colSpan="4">No orders</td></tr>}
          </tbody>
        </table>
      </div>
    </div>
  )
}

// --- Inventory Component ---
function InventoryPanel() {
  const [skuCheck, setSkuCheck] = useState('')
  const [inventoryData, setInventoryData] = useState(null)
  const [createForm, setCreateForm] = useState({
    product_id: 101,
    sku: "SKU-001",
    quantity: 100
  })

  const checkSku = async () => {
    if (!skuCheck) return
    try {
      // API Gateway maps /api/v1/products/:sku -> /inventory/sku?sku=:sku
      const res = await fetch(`/api/v1/products/${skuCheck}`)
      if (!res.ok) throw new Error('Not found or error')
      const data = await res.json()
      setInventoryData(data.data)
    } catch (err) {
      setInventoryData(null)
      alert(err.message)
    }
  }

  const createInventory = async () => {
    try {
      // API Gateway maps /api/v1/inventory/create -> /inventory/create
      const res = await fetch('/api/v1/inventory/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          product_id: Number(createForm.product_id),
          sku: createForm.sku,
          quantity: Number(createForm.quantity)
        })
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.msg || 'Error creating inventory')
      alert('Inventory Created!')
    } catch (err) {
      alert(err.message)
    }
  }

  return (
    <div className="panel">
      <div className="card">
        <h2>Create/Update Stock</h2>
        <div className="form-grid">
          <label>
            Product ID:
            <input type="number" value={createForm.product_id} onChange={e => setCreateForm({ ...createForm, product_id: e.target.value })} />
          </label>
          <label>
            SKU:
            <input type="text" value={createForm.sku} onChange={e => setCreateForm({ ...createForm, sku: e.target.value })} />
          </label>
          <label>
            Initial Qty:
            <input type="number" value={createForm.quantity} onChange={e => setCreateForm({ ...createForm, quantity: e.target.value })} />
          </label>
          <button onClick={createInventory}>Create Inventory</button>
        </div>
      </div>

      <div className="card">
        <h2>Check Stock</h2>
        <div className="search-bar">
          <input
            type="text"
            placeholder="Enter SKU (e.g. SKU-001)"
            value={skuCheck}
            onChange={e => setSkuCheck(e.target.value)}
          />
          <button onClick={checkSku}>Check</button>
        </div>
        {inventoryData && (
          <div className="result-box">
            <p><strong>SKU:</strong> {inventoryData.sku}</p>
            <p><strong>Quantity:</strong> {inventoryData.quantity}</p>
            <p><strong>Product ID:</strong> {inventoryData.product_id}</p>
          </div>
        )}
      </div>
    </div>
  )
}

// --- Payment Component ---
function PaymentsPanel() {
  const [orderIdCheck, setOrderIdCheck] = useState('')
  const [paymentData, setPaymentData] = useState(null)
  const [createForm, setCreateForm] = useState({
    order_id: "",
    amount: 1000
  })

  const checkPayment = async () => {
    if (!orderIdCheck) return
    try {
      const res = await fetch(`/api/v1/payments?order_id=${orderIdCheck}`)
      if (!res.ok) throw new Error('Not found')
      const data = await res.json()
      setPaymentData(data.data)
    } catch (err) {
      setPaymentData(null)
      alert(err.message)
    }
  }

  const createPayment = async () => {
    try {
      const res = await fetch('/api/v1/payments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          order_id: createForm.order_id,
          amount: Number(createForm.amount)
        })
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.msg || 'Error processing payment')
      alert(`Payment Processed! Status: ${data.data.status}`)
      setOrderIdCheck(createForm.order_id)
      checkPayment()
    } catch (err) {
      alert(err.message)
    }
  }

  return (
    <div className="panel">
      <div className="card">
        <h2>Process Payment</h2>
        <div className="form-grid">
          <label>
            Order ID:
            <input type="text" value={createForm.order_id} onChange={e => setCreateForm({ ...createForm, order_id: e.target.value })} />
          </label>
          <label>
            Amount (cents):
            <input type="number" value={createForm.amount} onChange={e => setCreateForm({ ...createForm, amount: e.target.value })} />
          </label>
          <button onClick={createPayment}>Pay</button>
        </div>
      </div>

      <div className="card">
        <h2>Check Payment Status</h2>
        <div className="search-bar">
          <input
            type="text"
            placeholder="Order ID"
            value={orderIdCheck}
            onChange={e => setOrderIdCheck(e.target.value)}
          />
          <button onClick={checkPayment}>Check</button>
        </div>
        {paymentData && (
          <div className="result-box">
            <p><strong>Status:</strong> <span className={`status ${paymentData.status}`}>{paymentData.status}</span></p>
            <p><strong>Amount:</strong> ${(paymentData.amount / 100).toFixed(2)}</p>
            <p><strong>Trans ID:</strong> {paymentData.transaction_id || '-'}</p>
          </div>
        )}
      </div>
    </div>
  )
}

export default App
