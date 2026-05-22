import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api, ApiError, type Candidate } from '../api/candidates'
import CandidateForm from '../components/CandidateForm'

export default function CandidateEdit() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const [candidate, setCandidate] = useState<Candidate | null>(null)
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [serverErrors, setServerErrors] = useState<string[]>([])

  useEffect(() => {
    if (!id) return
    api.get(Number(id))
      .then(setCandidate)
      .catch((e: Error) => setLoadError(e.message))
      .finally(() => setLoading(false))
  }, [id])

  async function handleSubmit(data: Omit<Candidate, 'id'>) {
    if (!id) return
    try {
      const updated = await api.update(Number(id), data)
      navigate(`/candidates/${updated.id}`)
    } catch (e) {
      setServerErrors(e instanceof ApiError ? e.errors : ['Unexpected error. Please try again.'])
    }
  }

  if (loading) return <div className="state-message">Loading…</div>
  if (loadError) return <div className="state-message state-error">Error: {loadError}</div>
  if (!candidate) return null

  return (
    <div className="page">
      <div className="page-header">
        <Link to={`/candidates/${id}`} className="back-link">← Back to candidate</Link>
      </div>
      <h1>Edit Candidate</h1>
      <CandidateForm
        initialData={candidate}
        onSubmit={handleSubmit}
        submitLabel="Save Changes"
        serverErrors={serverErrors}
      />
    </div>
  )
}
