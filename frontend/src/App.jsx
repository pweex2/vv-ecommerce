import { useState, useEffect } from 'react'
import './App.css'

function App() {
  const [orders, setOrders] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  // Fetch orders
  const fetchOrders = async () => {
    setLoading(true)
    try {
      // 通过 /api/v1/orders 访问后端，Vite 代理或 Nginx 会处理转发
      const response = await fetch('/api/v1/orders')
      if (!response.ok) {
        throw new Error('Failed to fetch orders')
      }
      const data = await response.json()
      setOrders(data.data || []) // 假设后端返回结构是 { code: 0, data: [...] }
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  // Create a dummy order
  const createOrder = async () => {
    try {
      const response = await fetch('/api/v1/orders', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: 1, // 模拟用户ID
          items: [
            { product_id: "101", quantity: 1, price: 100 }
          ]
        }),
      })
      if (!response.ok) {
        const errorData = await response.json()
        alert(`Failed: ${errorData.msg || 'Unknown error'}`)
        return
      }
      alert('Order created successfully!')
      fetchOrders() // Refresh list
    } catch (err) {
      alert(`Error: ${err.message}`)
    }
  }

  useEffect(() => {
    fetchOrders()
  }, [])

  return (
    <div className="container">
      <h1>🛍️ E-Commerce Admin</h1>
      
      <div className="card">
        <h2>Actions</h2>
        <button onClick={createOrder}>➕ Create Test Order</button>
        <button onClick={fetchOrders} style={{marginLeft: '10px'}}>🔄 Refresh List</button>
      </div>

      <div className="card">
        <h2>Order List</h2>
        {loading && <p>Loading...</p>}
        {error && <p style={{color: 'red'}}>Error: {error}</p>}
        
        {!loading && !error && (
          <table className="order-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>User ID</th>
                <th>Total Amount</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {orders.length === 0 ? (
                <tr>
                  <td colSpan="4">No orders found</td>
                </tr>
              ) : (
                orders.map(order => (
                  <tr key={order.id}>
                    <td>{order.id}</td>
                    <td>{order.user_id}</td>
                    <td>${(order.total_amount / 100).toFixed(2)}</td>
                    <td>{order.status}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}

export default App
