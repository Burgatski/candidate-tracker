import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api, ApiError, type Candidate } from '../api/candidates'
import CandidateForm from '../components/CandidateForm'

export default function CandidateCreate() {
  const navigate = useNavigate()
  const [serverErrors, setServerErrors] = useState<string[]>([])

  async function handleSubmit(data: Omit<Candidate, 'id'>) {
    try {
      const created = await api.create(data)
      navigate(`/candidates/${created.id}`)
    } catch (e) {
      setServerErrors(e instanceof ApiError ? e.errors : ['Unexpected error. Please try again.'])
    }
  }

  return (
    <div className="page">
      <div className="page-header">
        <Link to="/" className="back-link">← Back to list</Link>
      </div>
      <h1>Create Candidate</h1>
      <CandidateForm
        onSubmit={handleSubmit}
        submitLabel="Create Candidate"
        serverErrors={serverErrors}
      />
    </div>
  )
}
