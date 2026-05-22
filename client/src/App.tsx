import { BrowserRouter, Route, Routes } from 'react-router-dom'
import CandidateList from './pages/CandidateList'
import CandidateDetail from './pages/CandidateDetail'
import CandidateCreate from './pages/CandidateCreate'
import CandidateEdit from './pages/CandidateEdit'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<CandidateList />} />
        <Route path="/candidates/new" element={<CandidateCreate />} />
        <Route path="/candidates/:id" element={<CandidateDetail />} />
        <Route path="/candidates/:id/edit" element={<CandidateEdit />} />
      </Routes>
    </BrowserRouter>
  )
}
