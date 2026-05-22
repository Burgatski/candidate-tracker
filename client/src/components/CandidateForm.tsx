import { useState, type FormEvent, type ChangeEvent } from 'react'
import type { Candidate } from '../api/candidates'

interface Props {
  initialData?: Candidate
  onSubmit: (data: Omit<Candidate, 'id'>) => Promise<void>
  submitLabel: string
  serverErrors?: string[]
}

const COMMON_SKILLS = [
  'JavaScript', 'TypeScript', 'React', 'Vue', 'Angular',
  'Node.js', 'Python', 'Go', 'Java', 'Kotlin',
  'Swift', 'Rust', 'C#', 'PHP', 'Ruby',
  'SQL', 'PostgreSQL', 'MongoDB', 'Redis', 'Docker',
  'Kubernetes', 'AWS', 'GCP', 'Azure', 'GraphQL',
]

const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export default function CandidateForm({ initialData, onSubmit, submitLabel, serverErrors }: Props) {
  const [firstName, setFirstName] = useState(initialData?.first_name ?? '')
  const [lastName, setLastName] = useState(initialData?.last_name ?? '')
  const [email, setEmail] = useState(initialData?.email ?? '')
  const [phone, setPhone] = useState(initialData?.phone ?? '')
  const [picture, setPicture] = useState(initialData?.picture ?? '')
  const [picturePreview, setPicturePreview] = useState(initialData?.picture ?? '')
  const [skills, setSkills] = useState<string[]>(initialData?.skills ?? [])
  const [skillInput, setSkillInput] = useState('')
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [submitting, setSubmitting] = useState(false)

  function handleFileChange(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = () => {
      const result = reader.result as string
      setPicture(result)
      setPicturePreview(result)
    }
    reader.readAsDataURL(file)
  }

  function addSkill(value: string) {
    const trimmed = value.trim()
    if (trimmed && !skills.includes(trimmed)) {
      setSkills(prev => [...prev, trimmed])
    }
    setSkillInput('')
  }

  function removeSkill(skill: string) {
    setSkills(prev => prev.filter(s => s !== skill))
  }

  function validate(): boolean {
    const errs: Record<string, string> = {}
    if (!firstName.trim()) errs.first_name = 'First name is required'
    if (!lastName.trim()) errs.last_name = 'Last name is required'
    if (!phone.trim()) errs.phone = 'Phone is required'
    if (!picture) errs.picture = 'Picture is required'
    if (!emailRegex.test(email)) errs.email = 'Valid email is required'
    setErrors(errs)
    return Object.keys(errs).length === 0
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (!validate()) return
    setSubmitting(true)
    try {
      await onSubmit({ first_name: firstName, last_name: lastName, email, phone, picture, skills })
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="candidate-form">
      <div className="form-row">
        <div className="form-group">
          <label htmlFor="first_name">First Name *</label>
          <input
            id="first_name"
            type="text"
            value={firstName}
            onChange={e => setFirstName(e.target.value)}
            className={errors.first_name ? 'input-error' : ''}
          />
          {errors.first_name && <span className="error-msg">{errors.first_name}</span>}
        </div>
        <div className="form-group">
          <label htmlFor="last_name">Last Name *</label>
          <input
            id="last_name"
            type="text"
            value={lastName}
            onChange={e => setLastName(e.target.value)}
            className={errors.last_name ? 'input-error' : ''}
          />
          {errors.last_name && <span className="error-msg">{errors.last_name}</span>}
        </div>
      </div>

      <div className="form-row">
        <div className="form-group">
          <label htmlFor="email">Email *</label>
          <input
            id="email"
            type="text"
            value={email}
            onChange={e => setEmail(e.target.value)}
            className={errors.email ? 'input-error' : ''}
          />
          {errors.email && <span className="error-msg">{errors.email}</span>}
        </div>
        <div className="form-group">
          <label htmlFor="phone">Phone *</label>
          <input
            id="phone"
            type="tel"
            value={phone}
            onChange={e => setPhone(e.target.value)}
            className={errors.phone ? 'input-error' : ''}
          />
          {errors.phone && <span className="error-msg">{errors.phone}</span>}
        </div>
      </div>

      <div className="form-group">
        <label htmlFor="picture">Picture *</label>
        <input
          id="picture"
          type="file"
          accept="image/*"
          onChange={handleFileChange}
          className={errors.picture ? 'input-error' : ''}
        />
        {errors.picture && <span className="error-msg">{errors.picture}</span>}
        {picturePreview && (
          <img src={picturePreview} alt="Preview" className="picture-preview" />
        )}
      </div>

      <div className="form-group">
        <label htmlFor="skill-input">Skills</label>
        <div className="skills-input">
          <input
            id="skill-input"
            type="text"
            value={skillInput}
            onChange={e => setSkillInput(e.target.value)}
            onKeyDown={e => {
              if (e.key === 'Enter') {
                e.preventDefault()
                addSkill(skillInput)
              }
            }}
            placeholder="Type a skill and press Enter…"
            list="skills-datalist"
          />
          <datalist id="skills-datalist">
            {COMMON_SKILLS.map(s => <option key={s} value={s} />)}
          </datalist>
          <button type="button" className="btn-secondary" onClick={() => addSkill(skillInput)}>
            Add
          </button>
        </div>
        {skills.length > 0 && (
          <div className="skills-tags">
            {skills.map(skill => (
              <span key={skill} className="skill-tag">
                {skill}
                <button type="button" onClick={() => removeSkill(skill)} aria-label={`Remove ${skill}`}>
                  ×
                </button>
              </span>
            ))}
          </div>
        )}
      </div>

      {serverErrors && serverErrors.length > 0 && (
        <div className="server-errors">
          {serverErrors.map((err, i) => <p key={i}>{err}</p>)}
        </div>
      )}

      <button type="submit" disabled={submitting} className="btn-primary">
        {submitting ? 'Saving…' : submitLabel}
      </button>
    </form>
  )
}
