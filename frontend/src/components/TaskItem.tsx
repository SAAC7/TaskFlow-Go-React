import { useState } from 'react'
import { getErrorMessage } from '../api'
import type { Task, TaskPatch } from '../types'

interface Props {
  task: Task
  onUpdate: (id: string, patch: TaskPatch) => Promise<void>
  onDelete: (id: string) => Promise<void>
}

export function TaskItem({ task, onUpdate, onDelete }: Props) {
  const [editing, setEditing] = useState(false)
  const [draftTitle, setDraftTitle] = useState(task.title)
  const [draftDescription, setDraftDescription] = useState(task.description)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Envuelve cualquier acción: marca "ocupado", captura el error y lo muestra.
  async function run(action: () => Promise<void>) {
    setBusy(true)
    setError(null)
    try {
      await action()
    } catch (err) {
      setError(getErrorMessage(err))
    } finally {
      setBusy(false)
    }
  }

  // Al entrar a editar, partimos siempre de los valores actuales de la tarea.
  function startEditing() {
    setDraftTitle(task.title)
    setDraftDescription(task.description)
    setError(null)
    setEditing(true)
  }

  function cancelEditing() {
    setEditing(false)
    setError(null)
  }

  const toggle = () => run(() => onUpdate(task.id, { completed: !task.completed }))
  const remove = () => run(() => onDelete(task.id))
  const save = () =>
    run(async () => {
      if (!draftTitle.trim()) {
        throw new Error('El título es obligatorio')
      }
      await onUpdate(task.id, { title: draftTitle, description: draftDescription })
      setEditing(false)
    })

  return (
    <li className={`task ${task.completed ? 'done' : ''}`}>
      <input
        type="checkbox"
        checked={task.completed}
        onChange={toggle}
        disabled={busy || editing}
        aria-label="Marcar como completada"
      />

      <div className="task-body">
        {editing ? (
          <div className="task-edit">
            <input
              value={draftTitle}
              onChange={(e) => setDraftTitle(e.target.value)}
              maxLength={120}
              autoFocus
              disabled={busy}
              aria-label="Editar título"
            />
            <textarea
              value={draftDescription}
              onChange={(e) => setDraftDescription(e.target.value)}
              maxLength={1000}
              rows={3}
              disabled={busy}
              placeholder="Descripción (opcional)"
              aria-label="Editar descripción"
            />
          </div>
        ) : (
          <>
            <span className="task-title">{task.title}</span>
            {task.description && <p className="task-desc">{task.description}</p>}
            <small>{new Date(task.createdAt).toLocaleDateString('es-GT')}</small>
          </>
        )}
        {error && <p className="error">{error}</p>}
      </div>

      <div className="task-actions">
        {editing ? (
          <>
            <button onClick={save} disabled={busy}>Save</button>
            <button onClick={cancelEditing} disabled={busy}>Cancel</button>
          </>
        ) : (
          <>
            <button onClick={startEditing} disabled={busy}>Edit</button>
            <button className="danger" onClick={remove} disabled={busy}>Delete</button>
          </>
        )}
      </div>
    </li>
  )
}