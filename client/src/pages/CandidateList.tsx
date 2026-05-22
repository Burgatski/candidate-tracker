import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { api, type CandidateListData } from '../api/candidates'
import Pagination from '../components/Pagination'

export default function CandidateList() {
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Math.max(1, parseInt(searchParams.get('page') ?? '1', 10))

  const [data, setData] = useState<CandidateListData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    setError(null)
    api.list(page)
      .then(setData)
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false))
  }, [page])

  if (loading) return <div className="state-message">Loading candidates…</div>
  if (error) return <div className="state-message state-error">Error: {error}</div>
  if (!data) return null

  return (
    <div className="page">
      <div className="page-header">
        <h1>Candidates <span className="total-count">({data.total})</span></h1>
        <Link to="/candidates/new" className="btn-primary">+ Create Candidate</Link>
      </div>

      <div className="table-container">
        <table className="candidates-table">
          <thead>
            <tr>
              <th>Photo</th>
              <th>Name</th>
              <th>Email</th>
              <th>Skills</th>
            </tr>
          </thead>
          <tbody>
            {data.candidates.map(c => (
              <tr key={c.id}>
                <td>
                  <img
                    src={c.picture}
                    alt={`${c.first_name} ${c.last_name}`}
                    className="candidate-avatar"
                  />
                </td>
                <td>
                  <Link to={`/candidates/${c.id}`} className="candidate-link">
                    {c.first_name} {c.last_name}
                  </Link>
                </td>
                <td className="text-muted">{c.email}</td>
                <td>
                  <div className="skills-list">
                    {(c.skills ?? []).slice(0, 4).map(skill => (
                      <span key={skill} className="skill-badge">{skill}</span>
                    ))}
                    {(c.skills ?? []).length > 4 && (
                      <span className="skill-badge skill-badge-more">
                        +{c.skills.length - 4}
                      </span>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <Pagination
        page={data.pagination.page}
        totalPages={data.pagination.total_pages}
        onPageChange={newPage => setSearchParams({ page: String(newPage) })}
      />
    </div>
  )
}
