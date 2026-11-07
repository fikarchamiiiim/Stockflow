import { BrowserRouter, Route, Routes } from 'react-router-dom'
import OrderDetailPage from './pages/OrderDetailPage'
import ProductsPage from './pages/ProductsPage'

function App() {
  return (
    <BrowserRouter>
      <main>
        <Routes>
          <Route path="/" element={<ProductsPage />} />
          <Route path="/orders/:id" element={<OrderDetailPage />} />
        </Routes>
      </main>
    </BrowserRouter>
  )
}

export default App
