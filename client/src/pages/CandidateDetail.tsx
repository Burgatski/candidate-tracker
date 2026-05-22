import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api, type Candidate } from '../api/candidates'
import ConfirmDialog from '../components/ConfirmDialog'

export default function CandidateDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const [candidate, setCandidate] = useState<Candidate | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showConfirm, setShowConfirm] = useState(false)
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    if (!id) return
    api.get(Number(id))
      .then(setCandidate)
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false))
  }, [id])

  async function handleDelete() {
    if (!id) return
    setDeleting(true)
    try {
      await api.delete(Number(id))
      navigate('/')
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Delete failed')
      setShowConfirm(false)
      setDeleting(false)
    }
  }

  if (loading) return <div className="state-message">Loading…</div>
  if (error && !candidate) return <div className="state-message state-error">Error: {error}</div>
  if (!candidate) return null

  return (
    <div className="page">
      <div className="page-header">
        <Link to="/" className="back-link">← Back to list</Link>
        <div className="header-actions">
          <Link to={`/candidates/${candidate.id}/edit`} className="btn-secondary">
            Edit
          </Link>
          <button
            className="btn-danger"
            onClick={() => setShowConfirm(true)}
            disabled={deleting}
          >
            {deleting ? 'Deleting…' : 'Delete'}
          </button>
        </div>
      </div>

      <div className="detail-card">
        <div className="detail-left">
          <img
            src={candidate.picture}
            alt={`${candidate.first_name} ${candidate.last_name}`}
            className="candidate-photo"
          />
        </div>
        <div className="detail-right">
          <h1>{candidate.first_name} {candidate.last_name}</h1>

          <div className="detail-field">
            <span className="detail-label">Email</span>
            <span>{candidate.email}</span>
          </div>
          <div className="detail-field">
            <span className="detail-label">Phone</span>
            <span>{candidate.phone}</span>
          </div>
          <div className="detail-field">
            <span className="detail-label">Skills</span>
            <div className="skills-list">
              {(candidate.skills ?? []).length > 0
                ? candidate.skills.map(skill => (
                    <span key={skill} className="skill-badge">{skill}</span>
                  ))
                : <span className="text-muted">No skills listed</span>
              }
            </div>
          </div>
        </div>
      </div>

      {error && <div className="server-errors"><p>{error}</p></div>}

      {showConfirm && (
        <ConfirmDialog
          message={`Delete ${candidate.first_name} ${candidate.last_name}? This action cannot be undone.`}
          onConfirm={handleDelete}
          onCancel={() => setShowConfirm(false)}
        />
      )}
    </div>
  )
}
